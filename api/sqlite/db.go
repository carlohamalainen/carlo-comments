package sqlite

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
)

type DB struct {
	*sql.DB
}

func Open(ctx context.Context, cfg config.Config) (*DB, error) {
	logger := conduit.GetLogger(ctx)

	var db *sql.DB
	var err error

	db, err = sql.Open("sqlite3", cfg.SqlitePath)
	if err != nil {
		logger.Error("could not open database", "error", err)
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			comment_id TEXT PRIMARY KEY,
			site_id TEXT NOT NULL,
			post_id TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			author TEXT NOT NULL,
			author_email TEXT NOT NULL,
			comment TEXT NOT NULL,
			is_active INTEGER CHECK (is_active IN (0, 1))
		);
    `)
	if err != nil {
		logger.Error("failed to exec CREATE TABLE for comments", "error", err)
		return nil, err
	}

	return &DB{db}, nil
}
