package db

import (
	"fmt"
	"review-service/config"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var connectDB = sqlx.Connect

// Open creates the review-service YugabyteDB connection pool.
func Open(cfg config.DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Review,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := connectDB("pgx", dsn)
	if err != nil {
		return nil, AnnotateOpenDBError(err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(2 * time.Minute)

	return db, nil
}
