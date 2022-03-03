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
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Those flags define the MQTT Quality of Service (QoS) levels
const (
	MQTT_QOS_0 = 0 // QoS 1
	MQTT_QOS_1 = 1 // QoS 2
	MQTT_QOS_2 = 2 // QoS 3
)

// An MqttConfig represents the required information to connect to an MQTT
// broker.
type MqttConfig struct {
	BrokerURL   string        // broker url (tcp://hostname:port or ssl://hostname:port)
	Username    string        // username (optional)
	Password    string        // password (optional)
	ClientID    string        // MQTT ClientID
	Timeout     time.Duration // how much time to wait for connect and subscribe operations to complete
	GracePeriod time.Duration // how much time to wait for the disconnect operation to complete
}

// SetMqttLogger sets the logger to be used by the underlying MQTT library
func SetMqttLogger(logger *log.Logger) {
	mqtt.CRITICAL = logger
	mqtt.ERROR = logger
	mqtt.WARN = logger
}

// NewMqttClient creates a new MQTT client and connects to the broker
func NewMqttClient(config MqttConfig) (mqtt.Client, error) {
	if config.BrokerURL == "" {
		return nil, fmt.Errorf("MQTT broker URL is empty")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.BrokerURL)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(config.Timeout)
	opts.SetOrderMatters(false)
	opts.SetCleanSession(false)
	opts.SetClientID(config.ClientID)
	if config.Username != "" {
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
	}

	client := mqtt.NewClient(opts)
	ct := client.Connect()
	if !ct.WaitTimeout(config.Timeout) {
		return nil, fmt.Errorf("mqtt: timeout waiting for connection")
	}

	return client, nil
}
