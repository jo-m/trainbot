package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/stitch"
)

const dbTsFormat = "2006-01-02 15:04:05.999Z07:00"

// InsertTrain inserts a new train sighting into the database.
// Returns the db id of the new row.
func InsertTrain(db *sqlx.DB, t stitch.Train) (int64, error) {
	var id int64
	const q = `
	INSERT INTO trains_v2 (
		start_ts,
		n_frames,
		length_px,
		speed_px_s,
		accel_px_s_2,
		px_per_m
	)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING id;`
	err := db.Get(&id, q,
		t.StartTS.Format(dbTsFormat),
		t.NFrames,
		t.LengthPx,
		t.SpeedPxS,
		t.AccelPxS2,
		t.Conf.PixelsPerM)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// This should have been ".000_-07:00"... but it's too late now.
const fileTSFormat = "20060102_150405.999_Z07:00"

type Train struct {
	ID      int64     `db:"id"`
	StartTS time.Time `db:"start_ts"`
}

func (t *Train) GIFFileName() string {
	tsString := t.StartTS.Format(fileTSFormat)
	return fmt.Sprintf("train_%s.gif", tsString)
}

func (t *Train) ImgFileName() string {
	tsString := t.StartTS.Format(fileTSFormat)
	return fmt.Sprintf("train_%s.jpg", tsString)
}

// GetNextUpload returns the next train sighting to upload from the database.
func GetNextUpload(db *sqlx.DB) (*Train, error) {
	const q = `
	SELECT
		id, start_ts
	FROM trains_v2
	WHERE NOT uploaded
	ORDER BY id ASC
	LIMIT 1;
	`

	ret := Train{}
	err := db.Get(&ret, q)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

var ErrNoRowAffected = errors.New("no rows affected")

// SetUploaded marks a train sighting as uploaded in the database.
func SetUploaded(db *sqlx.DB, id int64) error {
	const q = `
	UPDATE trains_v2
	SET uploaded = TRUE
	WHERE id = ? AND NOT uploaded;
	`
	res, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return ErrNoRowAffected
	}

	return err
}

// GetNextCleanup returns the next train sighting for which we can delete the blobs locally.
func GetNextCleanup(db *sqlx.DB) (*Train, error) {
	const keepLastN = 100

	const q = `
	SELECT
		id, start_ts
	FROM trains_v2
	WHERE
		uploaded
		AND NOT cleaned_up
	ORDER BY id DESC
	LIMIT 1
	-- Always keep n last blobs.
	OFFSET ?;
	`

	ret := Train{}
	err := db.Get(&ret, q, keepLastN-1)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// SetCleanedUp marks a train sighting as uploaded in the database.
func SetCleanedUp(db *sqlx.DB, id int64) error {
	const q = `
	UPDATE trains_v2
	SET cleaned_up = TRUE
	WHERE id = ? AND NOT cleaned_up;
	`
	res, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return ErrNoRowAffected
	}

	return err
}

// InsertTemp inserts a new temperature measurement.
// Returns the db id of the new row.
func InsertTemp(db *sqlx.DB, ts time.Time, tempDegC float64) (int64, error) {
	var id int64
	const q = `
	INSERT INTO temperatures (
		timestamp,
		temp_deg_c
	)
	VALUES (?, ?)
	RETURNING id;`
	err := db.Get(&id, q,
		ts.Format(dbTsFormat), tempDegC)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAllBlobs lists all blobs which the database knows about.
// Does not include thumbnails.
func GetAllBlobs(db *sqlx.DB) (map[string]struct{}, error) {
	const q = `
	SELECT
		id, start_ts
	FROM trains_v2;`

	rows, err := db.Queryx(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string]struct{})
	var train Train
	for rows.Next() {
		err := rows.StructScan(&train)
		if err != nil {
			return nil, err
		}
		ret[train.ImgFileName()] = struct{}{}
		ret[train.GIFFileName()] = struct{}{}
	}

	return ret, nil
}
