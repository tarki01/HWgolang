package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

func Run(db *sql.DB, log *logrus.Logger) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("migrations table: %w", err)
	}

	files, err := fs.Glob(sqlFiles, "sql/*.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, path := range files {
		name := path[len("sql/"):]

		applied, err := isApplied(db, name)
		if err != nil {
			return err
		}
		if applied {
			log.WithField("migration", name).Debug("already applied, skipping")
			continue
		}

		content, err := sqlFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		if err := apply(db, name, string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}

		log.WithField("migration", name).Info("migration applied")
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

func isApplied(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name,
	).Scan(&exists)
	return exists, err
}

func apply(db *sql.DB, name, content string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range splitStatements(content) {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("statement failed: %w\nSQL: %s", err, stmt)
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (name) VALUES ($1)`, name,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func splitStatements(content string) []string {
	var stmts []string
	for _, s := range strings.Split(content, ";") {
		s = strings.TrimSpace(s)
		if s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}
