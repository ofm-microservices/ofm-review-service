package db

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"review-service/config"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/yugabytedb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type migrator interface {
	Up() error
}

var newMigrator = func(sourceURL, databaseURL string) (migrator, error) {
	return migrate.New(sourceURL, databaseURL)
}

// RunMigrations applies the review-service Yugabyte schema migrations.
func RunMigrations(cfg config.DBConfig) error {
	dsn := fmt.Sprintf(
		"yugabytedb://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		url.QueryEscape(cfg.Review),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		url.QueryEscape(cfg.MigrationsTable),
	)

	migrationPath := cfg.MigrationsPath
	if after, ok := strings.CutPrefix(migrationPath, "file://"); ok {
		raw := after
		if !filepath.IsAbs(raw) {
			abs, err := filepath.Abs(raw)
			if err != nil {
				return AnnotateResolveMigrationsPathError(err)
			}
			migrationPath = "file://" + abs
		}
	}

	m, err := newMigrator(migrationPath, dsn)
	if err != nil {
		return AnnotateCreateMigratorError(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return AnnotateRunMigrationsError(err)
	}

	return nil
}

// WithTx executes the callback within a SQL transaction and commits only if
// the callback succeeds.
func WithTx(ctx func(tx *sqlx.Tx) error, db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	if err := ctx(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
