/*
Copyright © 2022 Nicolas MASSE

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// An SqlConfig stores connection details to the database
type SqlConfig struct {
	Url string // Database URL (driver://user:password@hostname:port/db?opts)
}

// A ProcessorConfig stores the configuration of a processor
type ProcessorConfig struct {
	Sql    SqlConfig
	Mqtt   MqttConfig
	Logger *log.Logger
}

// A UnixEpoch is a time.Time that serializes / deserializes as Unix epoch
type UnixEpoch time.Time

// MarshalJSON returns the current value as JSON
func (t UnixEpoch) MarshalJSON() ([]byte, error) {
	t2 := time.Time(t)
	return []byte(fmt.Sprintf("%d", t2.Unix())), nil
}

// UnmarshalJSON initialises the current object from its JSON representation
func (t *UnixEpoch) UnmarshalJSON(b []byte) error {
	unix, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	*t = UnixEpoch(time.Unix(unix, 0))
	return nil
}

// A TicMessage represents data received from the TIC (Tele Information Client)
type TicMessage struct {
	Timestamp UnixEpoch `json:"ts"`
	Field     string    `json:"-"`
	Value     string    `json:"val"`
}

// A Processor receives events from the MQTT broker and saves data to the database
type Processor struct {
	Config   ProcessorConfig // the configuration
	client   mqtt.Client     // the MQTT client
	messages chan TicMessage // channel to send events from the MQTT go routines to the main method
	conn     *sql.DB         // the database connection
}

const (
	// How many in-flight MQTT messages to buffer
	MESSAGE_CHANNEL_LENGTH = 10

	// SQL Query to store current data
	UpsertCurrentQuery string = `
	INSERT INTO current VALUES ($1, $2, $3)
	ON CONFLICT (timestamp, phase) DO UPDATE
    SET current = excluded.current`

	// SQL Query to store power data
	UpsertPowerQuery string = `
	INSERT INTO power VALUES ($1, $2)
	ON CONFLICT (timestamp) DO UPDATE
    SET power = excluded.power`

	// SQL Query to store energy data
	UpsertEnergyQuery string = `
	INSERT INTO energy VALUES ($1, $2, $3)
	ON CONFLICT (timestamp, tariff) DO UPDATE
    SET reading = excluded.reading`
)

// NewProcessor creates a new processor from its configuration
func NewProcessor(c ProcessorConfig) *Processor {
	processor := Processor{
		Config:   c,
		messages: make(chan TicMessage, MESSAGE_CHANNEL_LENGTH),
	}
	return &processor
}

// usefulTopics is a list of topics of interest
var usefulTopics map[string]bool = map[string]bool{
	"IINST":  true,
	"IINST1": true,
	"IINST2": true,
	"IINST3": true,
	"PAPP":   true,
	"BASE":   true,
	"HCHP":   true,
	"HCHC":   true,
}

// Process receives MQTT messages and saves data to the SQL database
func (processor *Processor) Process() error {
	var err error

	// connect to the SQL Database
	processor.Config.Logger.Println("Connecting to PostgreSQL server...")
	processor.conn, err = sql.Open("pgx", processor.Config.Sql.Url)
	if err != nil {
		return err
	}
	defer processor.conn.Close()

	// do SQL Schema migrations
	processor.Config.Logger.Println("Ensuring db schema is up-to-date...")
	err = MigrateDb(processor.conn)
	if err != nil {
		return err
	}

	// connect to the MQTT broker
	SetMqttLogger(processor.Config.Logger)
	processor.Config.Logger.Println("Connecting to MQTT server...")
	processor.client, err = NewMqttClient(processor.Config.Mqtt)
	if err != nil {
		return err
	}

	// subscribe to topics
	topics := "esp-tic/status/tic/#"
	processor.Config.Logger.Printf("Subscribing to topics %s...", topics)
	st := processor.client.Subscribe(topics, MQTT_QOS_2, processor.processMessage)
	if !st.WaitTimeout(processor.Config.Mqtt.Timeout) {
		return fmt.Errorf("mqtt: timeout waiting for subscribe")
	}

	// process MQTT messages
	for {
		msg := <-processor.messages

		var err error
		if msg.Field == "IINST" || msg.Field == "IINST1" || msg.Field == "IINST2" || msg.Field == "IINST3" {
			err = processor.processCurrent(msg)
		} else if msg.Field == "PAPP" {
			err = processor.processPower(msg)
		} else if msg.Field == "BASE" || msg.Field == "HCHP" || msg.Field == "HCHC" {
			err = processor.processEnergy(msg)
		}

		if err != nil {
			processor.Config.Logger.Println(err)
		}
	}
}

// processCurrent saves current data to the database
func (processor *Processor) processCurrent(msg TicMessage) error {
	phase := 0
	if msg.Field != "IINST" {
		phase = int(msg.Field[5] - '0')
	}
	value, err := strconv.ParseInt(msg.Value, 10, 32)
	if err != nil {
		return err
	}

	rows, err := processor.conn.Query(UpsertCurrentQuery,
		time.Time(msg.Timestamp),
		phase,
		value)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

// processPower saves power data to the database
func (processor *Processor) processPower(msg TicMessage) error {
	value, err := strconv.ParseInt(msg.Value, 10, 32)
	if err != nil {
		return err
	}

	rows, err := processor.conn.Query(UpsertPowerQuery,
		time.Time(msg.Timestamp),
		value)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

// processEnergy saves energy readings to the database
func (processor *Processor) processEnergy(msg TicMessage) error {
	value, err := strconv.ParseInt(msg.Value, 10, 32)
	if err != nil {
		return err
	}

	rows, err := processor.conn.Query(UpsertEnergyQuery,
		time.Time(msg.Timestamp),
		msg.Field,
		value)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

// processMessage is the callback routine called by the MQTT library to process
// events.
func (processor *Processor) processMessage(c mqtt.Client, m mqtt.Message) {
	if m.Retained() {
		return
	}

	topic := m.Topic()
	pos := strings.LastIndexByte(topic, '/')
	if pos == -1 {
		return
	}

	field := topic[pos+1:]
	var ok bool
	if _, ok = usefulTopics[field]; !ok {
		return
	}

	var msg TicMessage
	err := json.Unmarshal(m.Payload(), &msg)
	if err != nil {
		processor.Config.Logger.Println(err)
		return
	}
	msg.Field = field

	processor.messages <- msg
}
