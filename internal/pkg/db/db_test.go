package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"jo-m.ch/go/trainbot/internal/pkg/stitch"
)

func Test_Open_Schema(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}

func Test_Backup(t *testing.T) {
	t0 = mustParseTime("2023-06-10T16:20:58.805+02:00")

	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")

	// Create DB.
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Insert row.
	id, err := InsertTrain(db, stitch.Train{
		StartTS: t0})
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Create backup.
	backupPath := filepath.Join(tmp, "test.db.bak")
	err = Backup(db, backupPath)
	assert.NoError(t, err)

	// Reopen backup.
	err = db.Close()
	assert.NoError(t, err)
	db, err = Open(backupPath)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Compare row data.
	next, err := GetNextUpload(db)
	assert.NoError(t, err)
	assert.Equal(t, id, next.ID)
	assert.Equal(t, t0, next.StartTS)
}
