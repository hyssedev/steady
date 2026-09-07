package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/hyssedev/steady/internal/monitor"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(ctx context.Context) (Database, error) {
	db, err := open(ctx)
	if err != nil {
		return Database{}, err
	}

	if err := createTables(ctx, db); err != nil {
		db.Close()
		return Database{}, err
	}

	return Database{
		DB: db,
	}, nil
}

func (db Database) SyncMonitors(ctx context.Context, monitors []monitor.Monitor) ([]monitor.Monitor, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)

	for i := range monitors {
		if err := tx.QueryRowContext(ctx, syncMonitorQuery, monitors[i].Name, monitors[i].URL.String(), now).
			Scan(&monitors[i].ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return monitors, nil
}

func (db Database) SaveCheck(
	ctx context.Context,
	monitorID int64,
	success bool,
	statusCode *int,
	latencyMs int64,
	checkErr error,
) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	successInt := 0
	if success {
		successInt = 1
	}

	var checkError any
	if checkErr != nil {
		checkError = checkErr.Error()
	}

	if _, err := db.DB.ExecContext(
		ctx,
		insertCheckQuery,
		monitorID,
		successInt,
		statusCode,
		latencyMs,
		checkError,
		now,
	); err != nil {
		return err
	}

	return nil
}

func open(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./steady.db")
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

func createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createTablesQuery)

	return err
}
