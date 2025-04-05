package db

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"jo-m.ch/go/trainbot/internal/pkg/stitch"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	t0 = mustParseTime("2023-06-10T16:20:58.805+02:00")
	t1 = mustParseTime("2023-06-10T16:21:05.982+02:00")
	t2 = mustParseTime("2023-11-10T12:57:45.897+01:00")
	t3 = mustParseTime("2023-11-10T13:57:49.289+00:00")
)

// This is more of a documentation than a test..
func Test_Train_Filenames(t *testing.T) {
	tr := Train{
		StartTS: mustParseTime("2023-12-24T09:58:52.660009478Z"),
	}
	assert.Equal(t, "train_20231224_095852.66_Z.jpg", tr.ImgFileName())
	assert.Equal(t, "train_20231224_095852.66_Z.gif", tr.GIFFileName())

	tr = Train{
		StartTS: mustParseTime("2023-12-24T11:19:12.839262415Z"),
	}
	assert.Equal(t, "train_20231224_111912.839_Z.jpg", tr.ImgFileName())
	assert.Equal(t, "train_20231224_111912.839_Z.gif", tr.GIFFileName())

	tr = Train{
		StartTS: mustParseTime("2023-10-28T17:31:50.709434526+01:00"),
	}
	assert.Equal(t, "train_20231028_173150.709_+01:00.jpg", tr.ImgFileName())
	assert.Equal(t, "train_20231028_173150.709_+01:00.gif", tr.GIFFileName())

	tr = Train{
		StartTS: mustParseTime("2023-11-25T15:49:46.958831882+00:00"),
	}
	assert.Equal(t, "train_20231125_154946.958_Z.jpg", tr.ImgFileName())
	assert.Equal(t, "train_20231125_154946.958_Z.gif", tr.GIFFileName())

	tr = Train{
		StartTS: mustParseTime("2023-03-28T06:32:16.516941205+01:00"),
	}
	assert.Equal(t, "train_20230328_063216.516_+01:00.jpg", tr.ImgFileName())
	assert.Equal(t, "train_20230328_063216.516_+01:00.gif", tr.GIFFileName())
}

func Test_Train_Queries(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Insert.
	id, err := InsertTrain(db, stitch.Train{StartTS: t0})
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Query upload.
	upl, err := GetNextUpload(db)
	assert.NoError(t, err)
	assert.Equal(t, t0, upl.StartTS)

	// Mark as uploaded.
	err = SetUploaded(db, upl.ID)
	assert.NoError(t, err)

	// Mark as uploaded again.
	err = SetUploaded(db, upl.ID)
	assert.Error(t, err, ErrNoRowAffected)

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
	_, err = InsertTrain(db, stitch.Train{StartTS: t1})
	assert.NoError(t, err)
	_, err = InsertTrain(db, stitch.Train{StartTS: t2})
	assert.NoError(t, err)
	_, err = InsertTrain(db, stitch.Train{StartTS: t3})
	assert.NoError(t, err)

	blobs, err := GetAllBlobs(db)
	assert.NoError(t, err)
	assert.Len(t, blobs, 8)
	assert.Contains(t, blobs, "train_20230610_162058.805_+02:00.gif")
	assert.Contains(t, blobs, "train_20230610_162058.805_+02:00.jpg")

	// Check cleanup query with positive results.
	for i := 0; i < 100; i++ {
		id, err := InsertTrain(db, stitch.Train{StartTS: t0.Add(time.Second * time.Duration(i+1))})
		require.NoError(t, err)

		err = SetUploaded(db, id)
		require.NoError(t, err)
	}
	cleanup, err := GetNextCleanup(db)
	assert.NoError(t, err)
	SetCleanedUp(db, cleanup.ID)
	err = SetCleanedUp(db, cleanup.ID)
	assert.Error(t, err, sql.ErrNoRows)
	_, err = GetNextCleanup(db)
	assert.Error(t, err, sql.ErrNoRows)
}

func Test_Train_TimesstampDBSerialization(t *testing.T) {
	tmp := t.TempDir()
	dbpath := filepath.Join(tmp, "test.db")
	db, err := Open(dbpath)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Insert.
	id, err := InsertTrain(db, stitch.Train{StartTS: t0})
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Check.
	var results []struct {
		StartTS string `db:"start_ts"`
	}
	err = db.Select(&results, "SELECT start_ts FROM trains_v2 ORDER BY id DESC")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "2023-06-10T16:20:58.805+02:00", results[0].StartTS)

	// Another round.
	id, err = InsertTrain(db, stitch.Train{StartTS: t2})
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	results = nil
	err = db.Select(&results, "SELECT start_ts FROM trains_v2 ORDER BY id DESC")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "2023-11-10T12:57:45.897+01:00", results[0].StartTS)
}
