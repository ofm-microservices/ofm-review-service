package db

import (
	"fmt"
	"review-service/config"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
)

var connectDB = sqlx.Connect

// Open creates the review-service PostgreSQL connection pool.
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

	driverName, err := otelsql.Register("pgx", otelsql.WithAttributes(attribute.String("db.system", "postgresql"), attribute.String("db.namespace", cfg.Name)))
	if err != nil {
		return nil, AnnotateOpenDBError(err)
	}
	db, err := connectDB(driverName, dsn)
	if err != nil {
		return nil, AnnotateOpenDBError(err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(2 * time.Minute)

	return db, nil
}
