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
package cmd

import (
	"os"

	ticTsdb "github.com/nmasse-itix/tic-tsdb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Saves MQTT events to TimescaleDB",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		ok := true
		if viper.GetString("sql.database") == "" {
			logger.Println("No database name defined in configuration")
			ok = false
		}
		if viper.GetString("sql.hostname") == "" {
			logger.Println("No database server defined in configuration")
			ok = false
		}
		if viper.GetString("mqtt.broker") == "" {
			logger.Println("No MQTT broker defined in configuration")
			ok = false
		}
		if !ok {
			logger.Println()
			cmd.Help()
			os.Exit(1)
		}

		logger.Println("Dispatching...")
		config := ticTsdb.ProcessorConfig{
			Sql: ticTsdb.SqlConfig{
				Url: getDatabaseUrl(),
			},
			Mqtt: ticTsdb.MqttConfig{
				BrokerURL:   viper.GetString("mqtt.broker"),
				Username:    viper.GetString("mqtt.username"),
				Password:    viper.GetString("mqtt.password"),
				ClientID:    viper.GetString("mqtt.clientId"),
				Timeout:     viper.GetDuration("mqtt.timeout"),
				GracePeriod: viper.GetDuration("mqtt.gracePeriod"),
			},
			Logger: logger,
		}
		processor := ticTsdb.NewProcessor(config)
		err := processor.Process()
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
}
