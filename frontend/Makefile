.PHONY: run format build

run:
	VITE_DB_URL=http://localhost:5173/data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/data/blobs \
		yarn run dev

run_local:
	VITE_DB_URL=http://localhost:5173/_data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/_data/blobs \
		yarn run dev

format:
	yarn run format

build:
	VITE_DB_URL=data/db.sqlite3 \
	VITE_BLOBS_URL=data/blobs \
		yarn run build
