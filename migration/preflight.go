// Package migration provides guarded utilities for moving legacy SQLite data
// to PostgreSQL. It intentionally does not write to either database.
package migration

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

// TableReport compares the number of rows for one application table.
type TableReport struct {
	Name            string `json:"name"`
	SourceRows      int64  `json:"source_rows"`
	DestinationRows int64  `json:"destination_rows"`
	PresentInSource bool   `json:"present_in_source"`
	PresentInTarget bool   `json:"present_in_target"`
}

// PreflightReport is safe to persist as migration evidence: it contains no
// rows, credentials, or connection strings.
type PreflightReport struct {
	SourceTableCount      int           `json:"source_table_count"`
	DestinationTableCount int           `json:"destination_table_count"`
	DestinationReady      bool          `json:"destination_ready"`
	Tables                []TableReport `json:"tables"`
}

// OpenSQLiteReadOnly opens a legacy SQLite file without granting this process
// write access. The caller remains responsible for importing the sqlite3
// database driver.
func OpenSQLiteReadOnly(path string) (*sql.DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving SQLite path: %w", err)
	}
	sourceURL := url.URL{Scheme: "file", Path: absPath, RawQuery: "mode=ro"}
	return sql.Open("sqlite3", sourceURL.String())
}

// Preflight compares the legacy SQLite source with an already-migrated
// PostgreSQL target. The target is ready only when every business table is
// empty. Goose bookkeeping is ignored because migrations create it first.
func Preflight(ctx context.Context, source, destination *sql.DB) (*PreflightReport, error) {
	sourceTables, err := sqliteTables(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("listing SQLite source tables: %w", err)
	}
	destinationTables, err := postgresTables(ctx, destination)
	if err != nil {
		return nil, fmt.Errorf("listing PostgreSQL destination tables: %w", err)
	}

	tableNames := make(map[string]struct{}, len(sourceTables)+len(destinationTables))
	for name := range sourceTables {
		tableNames[name] = struct{}{}
	}
	for name := range destinationTables {
		tableNames[name] = struct{}{}
	}

	report := &PreflightReport{
		SourceTableCount:      len(sourceTables),
		DestinationTableCount: len(destinationTables),
		DestinationReady:      true,
	}
	for name := range tableNames {
		_, inSource := sourceTables[name]
		_, inDestination := destinationTables[name]
		entry := TableReport{Name: name, PresentInSource: inSource, PresentInTarget: inDestination}
		if inSource {
			entry.SourceRows, err = countRows(ctx, source, name)
			if err != nil {
				return nil, fmt.Errorf("counting SQLite table %q: %w", name, err)
			}
		}
		if inDestination {
			entry.DestinationRows, err = countRows(ctx, destination, name)
			if err != nil {
				return nil, fmt.Errorf("counting PostgreSQL table %q: %w", name, err)
			}
			if !isMigrationMetadata(name) && entry.DestinationRows > 0 {
				report.DestinationReady = false
			}
		}
		report.Tables = append(report.Tables, entry)
	}
	sort.Slice(report.Tables, func(i, j int) bool { return report.Tables[i].Name < report.Tables[j].Name })
	return report, nil
}

func sqliteTables(ctx context.Context, database *sql.DB) (map[string]struct{}, error) {
	rows, err := database.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTableNames(rows)
}

func postgresTables(ctx context.Context, database *sql.DB) (map[string]struct{}, error) {
	rows, err := database.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTableNames(rows)
}

func scanTableNames(rows *sql.Rows) (map[string]struct{}, error) {
	tables := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables[name] = struct{}{}
	}
	return tables, rows.Err()
}

func countRows(ctx context.Context, database *sql.DB, table string) (int64, error) {
	var count int64
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoteIdentifier(table)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func isMigrationMetadata(table string) bool {
	return table == "goose_db_version"
}
