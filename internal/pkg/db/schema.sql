CREATE TABLE IF NOT EXISTS trains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    start_ts DATETIME NOT NULL UNIQUE,
    end_ts DATETIME NOT NULL,

    n_frames INT NOT NULL,
    length_px DOUBLE NOT NULL,
    speed_px_s DOUBLE NOT NULL,
    accel_px_s_2 DOUBLE NOT NULL,
    px_per_m  DOUBLE NOT NULL,

    -- relative to the blobs dir
    image_file_path TEXT NOT NULL UNIQUE,
    gif_file_path TEXT NOT NULL UNIQUE,

    -- if files from blob dir were uploaded
    uploaded_at DATETIME NULL DEFAULT NULL
);
