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

-- Periodic temperature measurements from trainbot compute hardware board.
-- Going to be interesting in summer.
CREATE TABLE IF NOT EXISTS temperatures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL UNIQUE,
    temp_deg_c DOUBLE NOT NULL
);

-- Schema v2!

CREATE TABLE IF NOT EXISTS trains_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    start_ts DATETIME NOT NULL UNIQUE,

    n_frames INT NOT NULL,
    -- Always positive (absolute value).
    length_px DOUBLE NOT NULL,
    -- Positive sign means movement to the right, negative to the left.
    speed_px_s DOUBLE NOT NULL,
    -- Positive sign means increasing speed for trains going to the right, breaking for trains going to the left.
    accel_px_s_2 DOUBLE NOT NULL,
    px_per_m  DOUBLE NOT NULL,

    -- Files from blob dir were uploaded.
    uploaded BOOL NOT NULL DEFAULT FALSE,

    -- Blobs we have deleted locally after upload.
    cleaned_up BOOL NOT NULL DEFAULT FALSE
);

BEGIN EXCLUSIVE TRANSACTION;

INSERT INTO trains_v2
SELECT
    id,
    start_ts,
    n_frames,
    length_px,
    speed_px_s,
    accel_px_s_2,
    px_per_m,
    uploaded_at IS NOT NULL,
    trains_blob_cleanups.cleaned_up_at IS NOT NULL
FROM trains
LEFT JOIN trains_blob_cleanups ON trains_blob_cleanups.train_id = trains.id
ORDER BY id ASC;

-- Truncate old tables.
DELETE FROM trains_blob_cleanups;
DELETE FROM trains;
DELETE FROM temperatures;

COMMIT;

-- Note: this SQL script will be run every time the database file is opened.
-- Any contents thus need to be idempotent (IF NOT EXISTS etc).
