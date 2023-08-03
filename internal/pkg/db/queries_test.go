package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/stitch"
	"github.com/stretchr/testify/assert"
)

func Test_Train_Queries(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// insert
	id, err := InsertTrain(db, stitch.Train{
		StartTS: time.Now(),
		EndTS:   time.Now(),
	}, "testimgpath", "gif")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// query upload
	upl, err := GetNextUpload(db)
	assert.NoError(t, err)
	assert.Equal(t, "testimgpath", upl.ImgPath)

	// mark as uploaded
	err = SetUploaded(db, upl.ID)
	assert.NoError(t, err)

	// query again
	_, err = GetNextUpload(db)
	assert.Equal(t, sql.ErrNoRows, err)

	// check cleanup queries - no results
	_, err = GetNextCleanup(db)
	assert.Equal(t, sql.ErrNoRows, err)

	err = SetCleanedUp(db, upl.ID)
	assert.NoError(t, err)
	err = SetCleanedUp(db, upl.ID)
	assert.Error(t, err)

	// test blobs listing query
	_, err = InsertTrain(db, stitch.Train{
		StartTS: time.Now(),
		EndTS:   time.Now(),
	}, "testimgpath2", "gif2")
	assert.NoError(t, err)

	blobs, err := GetAllBlobs(db)
	assert.NoError(t, err)
	assert.Len(t, blobs, 4)
	_, ok := blobs["testimgpath2"]
	assert.True(t, ok)
}

func Test_Temperature_Queries(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// insert
	id, err := InsertTemp(db, time.Now(), 123.456)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))
}
