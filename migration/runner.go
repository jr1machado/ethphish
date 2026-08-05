package migration

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Latest returns the highest numeric migration prefix in dir.
func Latest(dir string) (int64, error) {
	files, err := migrationFiles(dir)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, fmt.Errorf("no SQL migrations found in %s", dir)
	}
	return files[len(files)-1].version, nil
}

// Apply applies every SQL migration up to target. It supports the historical
// Goose annotation format but intentionally implements only the upward path
// used by the server. Rollbacks are rehearsed from backups, never by running
// down migrations against a live database.
func Apply(ctx context.Context, driver, dir string, target int64, database *sql.DB) error {
	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}
	if err := ensureVersionTable(ctx, driver, database); err != nil {
		return err
	}
	current, err := currentVersion(ctx, database)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.version <= current || file.version > target {
			continue
		}
		if err := applyFile(ctx, driver, database, file); err != nil {
			return err
		}
	}
	return nil
}

type migrationFile struct {
	version int64
	path    string
}

func migrationFiles(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			continue
		}
		version, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil || version <= 0 {
			continue
		}
		files = append(files, migrationFile{version: version, path: filepath.Join(dir, entry.Name())})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

func ensureVersionTable(ctx context.Context, driver string, database *sql.DB) error {
	var create, insert string
	switch driver {
	case "postgres":
		create = `CREATE TABLE IF NOT EXISTS goose_db_version (id BIGSERIAL PRIMARY KEY, version_id BIGINT NOT NULL, is_applied BOOLEAN NOT NULL, tstamp TIMESTAMP NULL DEFAULT now())`
		insert = `INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, $2)`
	case "sqlite3": // Legacy characterization tests only.
		create = `CREATE TABLE IF NOT EXISTS goose_db_version (id INTEGER PRIMARY KEY AUTOINCREMENT, version_id INTEGER NOT NULL, is_applied INTEGER NOT NULL, tstamp TIMESTAMP DEFAULT (datetime('now')))`
		insert = `INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?)`
	default:
		return fmt.Errorf("unsupported migration driver %q", driver)
	}
	if _, err := database.ExecContext(ctx, create); err != nil {
		return fmt.Errorf("creating migration version table: %w", err)
	}
	var count int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM goose_db_version`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := database.ExecContext(ctx, insert, 0, true); err != nil {
			return fmt.Errorf("initializing migration version table: %w", err)
		}
	}
	return nil
}

func currentVersion(ctx context.Context, database *sql.DB) (int64, error) {
	var version int64
	err := database.QueryRowContext(ctx, `SELECT version_id FROM goose_db_version WHERE is_applied = true ORDER BY id DESC LIMIT 1`).Scan(&version)
	return version, err
}

func applyFile(ctx context.Context, driver string, database *sql.DB, file migrationFile) error {
	f, err := os.Open(file.path)
	if err != nil {
		return err
	}
	defer f.Close()
	statements, err := splitUpSQL(f)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filepath.Base(file.path), err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying %s: %w", filepath.Base(file.path), err)
		}
	}
	insert := `INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, $2)`
	if driver == "sqlite3" {
		insert = `INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?)`
	}
	if _, err := tx.ExecContext(ctx, insert, file.version, true); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func splitUpSQL(r io.Reader) ([]string, error) {
	var statements []string
	var buffer bytes.Buffer
	scanner := bufio.NewScanner(r)
	active, block := false, false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "-- +goose ") {
			switch strings.TrimSpace(strings.TrimPrefix(line, "-- +goose ")) {
			case "Up":
				active = true
			case "Down":
				active = false
			case "StatementBegin":
				if active {
					block = true
				}
			case "StatementEnd":
				if active {
					block = false
					statements = append(statements, buffer.String())
					buffer.Reset()
				}
			}
		}
		if !active {
			continue
		}
		buffer.WriteString(line + "\n")
		if !block && sqlLineEnds(line) {
			statements = append(statements, buffer.String())
			buffer.Reset()
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(buffer.String()) != "" {
		return nil, fmt.Errorf("unterminated SQL statement")
	}
	return statements, nil
}

func sqlLineEnds(line string) bool {
	for _, word := range strings.Fields(line) {
		if strings.HasPrefix(word, "--") {
			break
		}
		if strings.HasSuffix(word, ";") {
			return true
		}
	}
	return false
}
