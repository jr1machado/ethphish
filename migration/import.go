package migration

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// ImportReport is the sanitized evidence emitted after a controlled import.
// It contains counts only, never source rows or connection strings.
type ImportReport struct {
	Preflight  *PreflightReport `json:"preflight"`
	Reconciled bool             `json:"reconciled"`
	Tables     []TableReport    `json:"tables"`
}

// Import copies compatible business tables from a read-only SQLite source into
// an empty, already-migrated PostgreSQL destination. It is transactional: a
// failed table rolls back every target write. Callers must obtain explicit
// change approval before invoking it.
func Import(ctx context.Context, source, destination *sql.DB) (*ImportReport, error) {
	preflight, err := Preflight(ctx, source, destination)
	if err != nil {
		return nil, err
	}
	if !preflight.DestinationReady {
		return nil, fmt.Errorf("PostgreSQL destination contains business data")
	}
	sourceTables, err := sqliteTables(ctx, source)
	if err != nil {
		return nil, err
	}
	targetTables, err := postgresTables(ctx, destination)
	if err != nil {
		return nil, err
	}
	for table := range sourceTables {
		if isMigrationMetadata(table) {
			continue
		}
		if _, ok := targetTables[table]; !ok {
			return nil, fmt.Errorf("source table %q is absent from PostgreSQL target", table)
		}
	}
	order, err := importOrder(ctx, destination, sourceTables)
	if err != nil {
		return nil, err
	}
	tx, err := destination.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, table := range order {
		if err := copyTable(ctx, source, tx, table); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := resetSequences(ctx, tx, table); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	report, err := Preflight(ctx, source, destination)
	if err != nil {
		return nil, err
	}
	reconciled := true
	for _, table := range report.Tables {
		if !isMigrationMetadata(table.Name) && table.PresentInSource && table.PresentInTarget && table.SourceRows != table.DestinationRows {
			reconciled = false
		}
	}
	return &ImportReport{Preflight: preflight, Reconciled: reconciled, Tables: report.Tables}, nil
}

func importOrder(ctx context.Context, destination *sql.DB, source map[string]struct{}) ([]string, error) {
	rows, err := destination.QueryContext(ctx, `SELECT tc.table_name, ccu.table_name FROM information_schema.table_constraints tc JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name AND tc.table_schema = ccu.table_schema WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	deps := map[string]map[string]bool{}
	for rows.Next() {
		var child, parent string
		if err := rows.Scan(&child, &parent); err != nil {
			return nil, err
		}
		_, childOK := source[child]
		_, parentOK := source[parent]
		if childOK && parentOK {
			if deps[child] == nil {
				deps[child] = map[string]bool{}
			}
			deps[child][parent] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var order []string
	remaining := map[string]bool{}
	for table := range source {
		if !isMigrationMetadata(table) {
			remaining[table] = true
		}
	}
	for len(remaining) > 0 {
		progressed := false
		names := make([]string, 0, len(remaining))
		for n := range remaining {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			ready := true
			for parent := range deps[n] {
				if remaining[parent] {
					ready = false
					break
				}
			}
			if ready {
				order = append(order, n)
				delete(remaining, n)
				progressed = true
			}
		}
		if !progressed {
			return nil, fmt.Errorf("cyclic or unresolved foreign-key dependencies in import")
		}
	}
	return order, nil
}

func copyTable(ctx context.Context, source *sql.DB, target *sql.Tx, table string) error {
	columns, err := sqliteColumns(ctx, source, table)
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return nil
	}
	quoted := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = quoteIdentifier(c)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	rows, err := source.QueryContext(ctx, "SELECT "+strings.Join(quoted, ",")+" FROM "+quoteIdentifier(table))
	if err != nil {
		return err
	}
	defer rows.Close()
	statement := "INSERT INTO " + quoteIdentifier(table) + " (" + strings.Join(quoted, ",") + ") VALUES (" + strings.Join(placeholders, ",") + ")"
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		if _, err := target.ExecContext(ctx, statement, values...); err != nil {
			return fmt.Errorf("inserting %s: %w", table, err)
		}
	}
	return rows.Err()
}

func sqliteColumns(ctx context.Context, database *sql.DB, table string) ([]string, error) {
	rows, err := database.QueryContext(ctx, "PRAGMA table_info("+quoteIdentifier(table)+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var def interface{}
		if err := rows.Scan(&cid, &name, &typ, &notnull, &def, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, rows.Err()
}

func resetSequences(ctx context.Context, target *sql.Tx, table string) error {
	rows, err := target.QueryContext(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 AND column_default LIKE 'nextval%'`, table)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return err
		}
		query := `SELECT setval(pg_get_serial_sequence($1,$2), GREATEST(COALESCE((SELECT MAX(` + quoteIdentifier(col) + `) FROM ` + quoteIdentifier(table) + `),1),1), true)`
		if _, err := target.ExecContext(ctx, query, table, col); err != nil {
			return err
		}
	}
	return rows.Err()
}
