.PHONY: run run_local format build deploy docker_build

# https://hub.docker.com/_/node/
DOCKER_BASE_IMAGE = node:20.7.0-alpine3.17
FRONTEND_DEPLOY_TARGET_SSH_HOST_ = ${FRONTEND_DEPLOY_TARGET_SSH_HOST}
VITE_FRONTEND_BASE_URL_ = ${VITE_FRONTEND_BASE_URL}

run:
	VITE_BASE_URL=http://localhost:5173/ \
	VITE_DB_URL=http://localhost:5173/data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/data/blobs \
		yarn run dev

run_local:
	# This requires some data in ../data.
	VITE_BASE_URL=http://localhost:5173/ \
	VITE_DB_URL=http://localhost:5173/_data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/_data/blobs \
		yarn run dev

format:
	yarn run format

build:
	test -n "$(VITE_FRONTEND_BASE_URL_)" # missing env var
	VITE_BASE_URL=$(VITE_FRONTEND_BASE_URL_) \
	VITE_DB_URL=$(VITE_FRONTEND_BASE_URL_)/data/db.sqlite3 \
	VITE_BLOBS_URL=$(VITE_FRONTEND_BASE_URL_)/data/blobs \
		yarn run build

deploy: build
	test -n "$(FRONTEND_DEPLOY_TARGET_SSH_HOST_)" # missing env var
	rm -rf dist/_data
	rsync --verbose --archive --compress --rsh=ssh dist/ $(FRONTEND_DEPLOY_TARGET_SSH_HOST_)

DOCKER_FLAGS = $(DOCKER_CLI_FLAGS)
DOCKER_FLAGS += --build-arg DOCKER_BASE_IMAGE="$(DOCKER_BASE_IMAGE)"

docker_build:
	docker build $(DOCKER_FLAGS)     \
		--target=export              \
		--output=dist                \
		.
