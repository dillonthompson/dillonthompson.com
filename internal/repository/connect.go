package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/dillonthompson/dillonthompson.com/internal/migrations"
	_ "github.com/lib/pq"
)

type Repository struct {
	DB      *sql.DB
	Queries *Queries
}

func Connect(databaseURL string) (*Repository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	const (
		maxRetries = 5
		baseDelay  = 2 * time.Second
	)

	for i := range maxRetries {
		if err = db.Ping(); err == nil {
			slog.Info("database connected")

			if err := migrations.Run(db); err != nil {
				db.Close()
				return nil, fmt.Errorf("running migrations: %w", err)
			}

			return &Repository{
				DB:      db,
				Queries: New(db),
			}, nil
		}
		delay := baseDelay * time.Duration(1<<i)
		slog.Warn("database not ready, retrying", "attempt", i+1, "delay", delay, "error", err)
		time.Sleep(delay)
	}

	db.Close()
	return nil, fmt.Errorf("database connection failed after %d retries: %w", maxRetries, err)
}
