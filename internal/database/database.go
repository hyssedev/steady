package database

import (
	"context"
	"database/sql"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(ctx context.Context) (Database, error) {
	db, err := open(ctx)
	if err != nil {
		return Database{}, err
	}

	return Database{
		DB: db,
	}, nil
}

func open(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3_extended", "./steady.db")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
