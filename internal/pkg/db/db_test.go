package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/stitch"
	"github.com/stretchr/testify/assert"
)

func Test_Open_Schema(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}

func Test_Backup(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")

	// create DB
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// insert row
	id, err := InsertTrain(db, stitch.Train{
		StartTS: time.Now(),
		EndTS:   time.Now(),
	}, "testimgpath", "gif")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// create backup
	backupPath := filepath.Join(tmp, "test.db.bak")
	err = Backup(db, backupPath)
	assert.NoError(t, err)

	// reopen backup
	err = db.Close()
	assert.NoError(t, err)
	db, err = Open(backupPath)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// compare row data
	next, err := GetNextUpload(db)
	assert.NoError(t, err)
	assert.Equal(t, "testimgpath", next.ImgPath)
}
