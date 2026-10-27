package store

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return conn, nil
}

func Init(conn *sql.DB) error {
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return fmt.Errorf("setting SQLite busy timeout: %w", err)
	}

	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("enabling SQLite WAL: %w", err)
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("reading schema: %w", err)
	}

	if _, err := conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}

	return nil
}
