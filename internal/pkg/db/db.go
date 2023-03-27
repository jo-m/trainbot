package db

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var sqlSchema string

const driver = "sqlite3"

func buildDSN(path string, readOnly bool) string {
	pragmas := map[string]string{
		"mode":          "rwc",
		"_journal_mode": "WAL",
		"_locking_mode": "NORMAL",
		"_txlock":       "deferred",
		"_foreign_keys": "true",
	}

	if readOnly {
		pragmas["mode"] = "ro"
	}

	path = path + "?"
	for k, v := range pragmas {
		path += k + "=" + v + "&"
	}

	return path[:len(path)-1]
}

// Open creates a new SQLite database, opens an existing one.
// Initializes with the schema if new.
func Open(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, buildDSN(path, false))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(sqlSchema)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, err
}
