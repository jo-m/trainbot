package db

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/stitch"
)

// Insert inserts a new train sighting into the database.
func Insert(db *sqlx.DB, t stitch.Train, imgPath, gifPath string) (int64, error) {
	var id int64
	const q = `
	INSERT INTO trains (
		start_ts,
		end_ts,
		n_frames,
		length_px,
		speed_px_s,
		accel_px_s_2,
		px_per_m,
		image_file_path,
		gif_file_path
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING id;`
	err := db.Get(&id, q,
		t.StartTS,
		t.EndTS,
		t.NFrames,
		t.LengthPx,
		t.SpeedPxS,
		t.AccelPxS2,
		t.Conf.PixelsPerM,
		imgPath,
		gifPath)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Upload represents a set of blobs for a train sighting.
type Upload struct {
	ID      int64  `db:"id"`
	ImgPath string `db:"image_file_path"`
	GIFPath string `db:"gif_file_path"`
}

// GetNextUpload returns the next train sighting to upload from the database.
func GetNextUpload(db *sqlx.DB) (*Upload, error) {
	const q = `
	SELECT
		id, image_file_path, gif_file_path
	FROM trains
	WHERE uploaded_at IS NULL
	ORDER BY id ASC
	LIMIT 1;
	`

	ret := Upload{}
	err := db.Get(&ret, q)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// SetUploaded marks a train sighting as uploaded in the database.
func SetUploaded(db *sqlx.DB, id int64) error {
	const q = `
	UPDATE trains SET uploaded_at = ? WHERE id = ?;
	`
	_, err := db.Exec(q, time.Now(), id)
	return err
}

// GetNextCleanup returns the next train sighting for which we can delete the blobs locally.
func GetNextCleanup(db *sqlx.DB) (*Upload, error) {
	const keepLastN = 100

	const q = `
	SELECT
		id, image_file_path, gif_file_path
	FROM trains
	LEFT JOIN trains_blob_cleanups
		ON trains_blob_cleanups.train_id = trains.id
	WHERE
		uploaded_at IS NOT NULL
		AND trains_blob_cleanups.cleaned_up_at IS NULL
	ORDER BY id DESC
	LIMIT 1
	-- Always keep n last blobs.
	OFFSET ?;
	`

	ret := Upload{}
	err := db.Get(&ret, q, keepLastN)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// SetCleanedUp marks a train sighting as uploaded in the database.
func SetCleanedUp(db *sqlx.DB, id int64) error {
	const q = `
	INSERT INTO trains_blob_cleanups (
		train_id,
		cleaned_up_at
	)
	VALUES(?, ?);
	`
	_, err := db.Exec(q, id, time.Now())
	return err
}
