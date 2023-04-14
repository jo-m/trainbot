package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
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
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, err
}

// Backup safely backs up a SQLite database to a new file.
func Backup(src *sqlx.DB, destPath string) error {
	err := os.Remove(destPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	dest, err := sqlx.Open(driver, destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	return doBackup(dest.DB, src.DB)
}

func doBackup(destDB, srcDB *sql.DB) error {
	destConn, err := destDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer destConn.Close()

	srcConn, err := srcDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer srcConn.Close()

	return destConn.Raw(func(destConn interface{}) error {
		return srcConn.Raw(func(srcConn interface{}) error {
			destSQLiteConn, ok := destConn.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("cannot convert destination connection to SQLiteConn")
			}

			srcSQLiteConn, ok := srcConn.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("cannot convert source connection to SQLiteConn")
			}

			b, err := destSQLiteConn.Backup("main", srcSQLiteConn, "main")
			if err != nil {
				return fmt.Errorf("failed to initialize backup: %w", err)
			}

			done, err := b.Step(-1)
			if !done {
				return errors.New("backup step -1, but not done")
			}
			if err != nil {
				return fmt.Errorf("failed to step backup: %w", err)
			}

			err = b.Finish()
			if err != nil {
				return fmt.Errorf("failed to finish backup: %w", err)
			}

			return err
		})
	})
}
