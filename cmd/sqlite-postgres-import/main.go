package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gophish/gophish/migration"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const approvalPhrase = "I_UNDERSTAND_THIS_WRITES_TO_POSTGRES"

func main() {
	var sourcePath, dsn, approval string
	flag.StringVar(&sourcePath, "sqlite", "", "legacy SQLite path")
	flag.StringVar(&dsn, "postgres-dsn", "", "PostgreSQL DSN")
	flag.StringVar(&approval, "approve", "", "required approval phrase")
	flag.Parse()
	if sourcePath == "" || dsn == "" || approval != approvalPhrase {
		fmt.Fprintln(os.Stderr, "usage: sqlite-postgres-import --sqlite <path> --postgres-dsn <dsn> --approve "+approvalPhrase)
		os.Exit(2)
	}
	source, err := migration.OpenSQLiteReadOnly(sourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening SQLite source:", err)
		os.Exit(1)
	}
	defer source.Close()
	target, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opening PostgreSQL destination:", err)
		os.Exit(1)
	}
	defer target.Close()
	report, err := migration.Import(context.Background(), source, target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "import failed:", err)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(report)
	if !report.Reconciled {
		os.Exit(3)
	}
}
