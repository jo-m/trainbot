package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/stitch"
)

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
	RETURNING id`
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
