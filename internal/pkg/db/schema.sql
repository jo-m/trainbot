-- Train sightings.
CREATE TABLE IF NOT EXISTS trains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    start_ts DATETIME NOT NULL UNIQUE,
    end_ts DATETIME NOT NULL,

    n_frames INT NOT NULL,
    -- Always positive (absolute value).
    length_px DOUBLE NOT NULL,
    -- Positive sign means movement to the right, negative to the left.
    speed_px_s DOUBLE NOT NULL,
    -- Positive sign means increasing speed for trains going to the right, breaking for trains going to the left.
    accel_px_s_2 DOUBLE NOT NULL,
    px_per_m  DOUBLE NOT NULL,

    -- Relative to the blobs dir.
    image_file_path TEXT NOT NULL UNIQUE,
    gif_file_path TEXT NOT NULL UNIQUE,

    -- Set if files from blob dir were uploaded.
    uploaded_at DATETIME NULL DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS trains_length ON trains(length_px / px_per_m);
CREATE INDEX IF NOT EXISTS trains_speed ON trains(ABS(speed_px_s / px_per_m));

-- Blobs we have deleted locally after upload.
CREATE TABLE IF NOT EXISTS trains_blob_cleanups (
    train_id INTEGER PRIMARY KEY,
    cleaned_up_at DATETIME NOT NULL,
    FOREIGN KEY(train_id) REFERENCES trains(id)
);
