package database

var createTablesQuery = `
		CREATE TABLE IF NOT EXISTS monitors (
			id         INTEGER PRIMARY KEY,
			name       TEXT NOT NULL,
			url        TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS checks (
			id          INTEGER PRIMARY KEY,
			monitor_id  INTEGER NOT NULL REFERENCES monitors(id),
			success     INTEGER NOT NULL,
			status_code INTEGER,
			latency_ms  INTEGER NOT NULL,
			error       TEXT,
			checked_at  TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS checks_monitor_checked_at
		ON checks (monitor_id, checked_at);`

var insertMonitorQuery = `
		INSERT INTO monitors (name, url, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET name = excluded.name;`

var getIdQuery = `
		SELECT id
		FROM monitors
		WHERE url = ?;`

var insertCheckQuery = `
		INSERT INTO checks (
			monitor_id,
			success,
			status_code,
			latency_ms,
			error,
			checked_at
		)
		VALUES (?, ?, ?, ?, ?, ?);`
