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

	for _, m := range monitors {
		if _, err := tx.ExecContext(
			ctx,
			insertMonitorQuery,
			m.Name,
			m.URL.String(),
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for i := range monitors {
		var id int64
		if err := db.DB.QueryRowContext(ctx, getIdQuery, monitors[i].URL.String()).Scan(&id); err != nil {
			return nil, err
		}

		monitors[i].ID = id
	}

	return monitors, nil
}

func (db Database) SaveCheck(
	ctx context.Context,
	monitorID int64,
	success bool,
	statusCode *int,
	latency_ms int64,
	checkErr error,
) error {
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
		latency_ms,
		checkError,
		time.Now().UTC().Format(time.RFC3339Nano),
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
