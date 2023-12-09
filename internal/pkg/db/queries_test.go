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

	// Insert.
	id, err := InsertTrain(db, stitch.Train{
		StartTS: time.Now(),
		EndTS:   time.Now(),
	}, "testimgpath", "gif")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Query upload.
	upl, err := GetNextUpload(db)
	assert.NoError(t, err)
	assert.Equal(t, "testimgpath", upl.ImgPath)

	// Mark as uploaded.
	err = SetUploaded(db, upl.ID)
	assert.NoError(t, err)

	// Query again.
	_, err = GetNextUpload(db)
	assert.Equal(t, sql.ErrNoRows, err)

	// Check cleanup queries - no results.
	_, err = GetNextCleanup(db)
	assert.Equal(t, sql.ErrNoRows, err)

	err = SetCleanedUp(db, upl.ID)
	assert.NoError(t, err)
	err = SetCleanedUp(db, upl.ID)
	assert.Error(t, err)

	// Test blobs listing query.
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

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	t0 = mustParseTime("2023-06-10T16:20:58.805488032+02:00")
	t1 = mustParseTime("2023-06-10T16:21:05.982318734+02:00")
	t2 = mustParseTime("2023-11-10T12:57:45.897342187+01:00")
	t3 = mustParseTime("2023-11-10T13:57:49.289348781+00:00")
)

func Test_Train_TimesstampDBSerialization(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Insert.
	id, err := InsertTrain(db, stitch.Train{
		StartTS: t0,
		EndTS:   t1,
	}, "testimgpath", "gif")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Check.
	var results []struct {
		StartTS string `db:"start_ts"`
		EndTS   string `db:"end_ts"`
	}
	err = db.Select(&results, "SELECT start_ts, end_ts FROM trains ORDER BY id DESC")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "2023-06-10T16:20:58.805488032+02:00", results[0].StartTS)
	assert.Equal(t, "2023-06-10T16:21:05.982318734+02:00", results[0].EndTS)

	// Another round.
	id, err = InsertTrain(db, stitch.Train{
		StartTS: t2,
		EndTS:   t3,
	}, "testimgpath2", "gif2")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	results = nil
	err = db.Select(&results, "SELECT start_ts, end_ts FROM trains ORDER BY id DESC")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "2023-11-10T12:57:45.897342187+01:00", results[0].StartTS)
	assert.Equal(t, "2023-11-10T13:57:49.289348781Z", results[0].EndTS)
}

func Test_Temperature_Queries(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Insert.
	id, err := InsertTemp(db, time.Now(), 123.456)
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))
}
