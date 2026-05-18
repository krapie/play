package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func Connect(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	conn.SetMaxOpenConns(1) // SQLite single-writer

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{conn: conn}, nil
}

func (d *DB) Conn() *sql.DB { return d.conn }

func (d *DB) Close() { d.conn.Close() }

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`
		PRAGMA journal_mode=WAL;

		CREATE TABLE IF NOT EXISTS results (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			game       TEXT    NOT NULL,
			player_id  TEXT    NOT NULL,
			value      INTEGER NOT NULL,
			meta       TEXT,
			created_at TEXT    DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_results_game_value
			ON results(game, value);

		CREATE INDEX IF NOT EXISTS idx_results_game_player
			ON results(game, player_id);

		CREATE TABLE IF NOT EXISTS button_presses (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL,
			pressed_at TEXT    DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_button_session ON button_presses(session_id);
		CREATE INDEX IF NOT EXISTS idx_button_time    ON button_presses(pressed_at);
	`)
	return err
}
