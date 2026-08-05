// postgres-schema-prepare applies the PostgreSQL schema without bootstrapping
// application users. It is intended only for isolated migration destinations.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gophish/gophish/migration"
	_ "github.com/lib/pq"
)

func main() {
	var dsn, migrations string
	flag.StringVar(&dsn, "postgres-dsn", "", "PostgreSQL destination DSN")
	flag.StringVar(&migrations, "migrations", "db/db_postgres/migrations", "PostgreSQL migrations directory")
	flag.Parse()
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "usage: postgres-schema-prepare --postgres-dsn <dsn>")
		os.Exit(2)
	}
	absMigrations, err := filepath.Abs(migrations)
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolving migrations path:", err)
		os.Exit(1)
	}
	latest, err := migration.Latest(absMigrations)
	if err != nil {
		fmt.Fprintln(os.Stderr, "finding latest migration:", err)
		os.Exit(1)
	}
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening PostgreSQL destination:", err)
		os.Exit(1)
	}
	defer database.Close()
	if err := migration.Apply(context.Background(), "postgres", absMigrations, latest, database); err != nil {
		fmt.Fprintln(os.Stderr, "applying migrations:", err)
		os.Exit(1)
	}
	fmt.Printf("PostgreSQL schema prepared at migration %d\n", latest)
}
