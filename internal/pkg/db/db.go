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

	sqlite3 "modernc.org/sqlite"
)

//go:embed schema.sql
var sqlSchema string

const driver = "sqlite"

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

// Open creates a new SQLite database or opens an existing one.
// Will run the schema/migration script.
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

	return doBackup(destPath, src.DB)
}

type backup interface {
	NewBackup(string) (*sqlite3.Backup, error)
}

func doBackup(destPath string, srcDB *sql.DB) error {
	srcConn, err := srcDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer srcConn.Close()

	return srcConn.Raw(func(srcConn interface{}) error {
		backup, ok := srcConn.(backup)
		if !ok {
			return errors.New("source connection does not implement NewBackup()")
		}

		bck, err := backup.NewBackup(destPath)
		if err != nil {
			return fmt.Errorf("failed to initialize backup: %w", err)
		}

		for more := true; more; {
			more, err = bck.Step(-1)
			if err != nil {
				return fmt.Errorf("failed to step backup: %w", err)
			}
		}

		err = bck.Finish()
		if err != nil {
			return fmt.Errorf("failed to finish backup: %w", err)
		}

		return nil
	})
}
