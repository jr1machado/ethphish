package models

// SQLite is registered only for the legacy characterization suite. The server
// binary imports PostgreSQL only; legacy SQLite files are otherwise opened by
// the dedicated migration command.
import _ "github.com/mattn/go-sqlite3"
