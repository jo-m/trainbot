.PHONY: run format build deploy

FRONTEND_DEPLOY_TARGET_SSH_HOST_ = ${FRONTEND_DEPLOY_TARGET_SSH_HOST}

run:
	VITE_DB_URL=http://localhost:5173/data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/data/blobs \
		yarn run dev

run_local:
	# This requires some data in ../data.
	VITE_DB_URL=http://localhost:5173/_data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/_data/blobs \
		yarn run dev

format:
	yarn run format

build:
	VITE_DB_URL=data/db.sqlite3 \
	VITE_BLOBS_URL=data/blobs \
		yarn run build

deploy: build
	rm -rf dist/_data
	rsync --verbose --archive --compress --rsh=ssh dist/ $(FRONTEND_DEPLOY_TARGET_SSH_HOST_)
