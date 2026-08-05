package migration

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestOpenSQLiteReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy source.db")
	writable, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Exec(`CREATE TABLE campaigns (id INTEGER PRIMARY KEY); INSERT INTO campaigns VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}

	readonly, err := OpenSQLiteReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer readonly.Close()
	var count int
	if err := readonly.QueryRow(`SELECT COUNT(*) FROM campaigns`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("row count = %d, want 1", count)
	}
	if _, err := readonly.Exec(`INSERT INTO campaigns VALUES (2)`); err == nil {
		t.Fatal("read-only source accepted a write")
	}
}

func TestSQLiteTablesAndCounts(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`CREATE TABLE users (id INTEGER); CREATE TABLE goose_db_version (id INTEGER); INSERT INTO users VALUES (1), (2)`); err != nil {
		t.Fatal(err)
	}

	tables, err := sqliteTables(context.Background(), database)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tables["users"]; !ok {
		t.Fatal("users table missing from SQLite inventory")
	}
	if _, ok := tables["goose_db_version"]; !ok {
		t.Fatal("migration metadata missing from SQLite inventory")
	}
	count, err := countRows(context.Background(), database, "users")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("users count = %d, want 2", count)
	}
}

func TestQuoteIdentifierAndMigrationMetadata(t *testing.T) {
	if got, want := quoteIdentifier(`odd"name`), `"odd""name"`; got != want {
		t.Fatalf("quoted identifier = %q, want %q", got, want)
	}
	if !isMigrationMetadata("goose_db_version") {
		t.Fatal("goose metadata should be ignored when checking target readiness")
	}
	if isMigrationMetadata("users") {
		t.Fatal("application table must not be ignored when checking target readiness")
	}
}
