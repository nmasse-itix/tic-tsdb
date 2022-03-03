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
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	ticTsdb "github.com/nmasse-itix/tic-tsdb"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database schema migrations",
	Long: `Goose commands:
up                   Migrate the DB to the most recent version available
up-by-one            Migrate the DB up by 1
up-to VERSION        Migrate the DB to a specific VERSION
down                 Roll back the version by 1
down-to VERSION      Roll back to a specific VERSION
redo                 Re-run the latest migration
reset                Roll back all migrations
status               Dump the migration status for the current DB
version              Print the current version of the database
create NAME [sql|go] Creates new migration file with the current timestamp
fix                  Apply sequential ordering to migrations
`,
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
		if len(args) < 1 {
			logger.Println("Please specify goose command!")
			ok = false
		}
		if !ok {
			logger.Println()
			cmd.Help()
			os.Exit(1)
		}

		dbUrl := getDatabaseUrl()
		logger.Println("Connecting to PostgreSQL server...")
		db, err := sql.Open("pgx", dbUrl)
		if err != nil {
			logger.Println(err)
			os.Exit(1)
		}
		defer db.Close()

		goose.SetBaseFS(ticTsdb.SqlMigrationFS)

		if err := goose.SetDialect("postgres"); err != nil {
			logger.Println(err)
			os.Exit(1)
		}

		gooseCmd := args[0]
		gooseOpts := args[1:]
		if err := goose.Run(gooseCmd, db, "schemas", gooseOpts...); err != nil {
			logger.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	dbCmd.AddCommand(migrateCmd)
}
