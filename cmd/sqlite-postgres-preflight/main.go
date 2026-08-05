// sqlite-postgres-preflight validates a legacy SQLite source before any data
// migration. It never writes to source or destination databases.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gophish/gophish/migration"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var sqlitePath string
	var postgresDSN string
	flag.StringVar(&sqlitePath, "sqlite", "", "path to the legacy SQLite database")
	flag.StringVar(&postgresDSN, "postgres-dsn", "", "PostgreSQL destination DSN")
	flag.Parse()
	if sqlitePath == "" || postgresDSN == "" {
		fmt.Fprintln(os.Stderr, "usage: sqlite-postgres-preflight --sqlite <path> --postgres-dsn <dsn>")
		os.Exit(2)
	}

	source, err := migration.OpenSQLiteReadOnly(sqlitePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening SQLite source:", err)
		os.Exit(1)
	}
	defer source.Close()
	destination, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening PostgreSQL destination:", err)
		os.Exit(1)
	}
	defer destination.Close()

	report, err := migration.Preflight(context.Background(), source, destination)
	if err != nil {
		fmt.Fprintln(os.Stderr, "preflight failed:", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, "writing report:", err)
		os.Exit(1)
	}
	if !report.DestinationReady {
		os.Exit(3)
	}
}
