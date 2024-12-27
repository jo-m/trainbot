.PHONY: run run_local format build deploy docker_build list

# https://hub.docker.com/_/node/
DOCKER_BASE_IMAGE = node:22.12.0-alpine
FRONTEND_DEPLOY_TARGET_SSH_HOST_ = ${FRONTEND_DEPLOY_TARGET_SSH_HOST}
VITE_FRONTEND_BASE_URL_ = ${VITE_FRONTEND_BASE_URL}

run:
	VITE_BASE_URL=http://localhost:5173/ \
	VITE_DB_URL=http://localhost:5173/data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/data/blobs \
		npm run dev

run_local:
	# This requires some data in ../data.
	VITE_BASE_URL=http://localhost:5173/ \
	VITE_DB_URL=http://localhost:5173/_data/db.sqlite3 \
	VITE_BLOBS_URL=http://localhost:5173/_data/blobs \
		npm run dev

format:
	npm run format

build:
	test -n "$(VITE_FRONTEND_BASE_URL_)" # missing env var, please set up env file from env.example and source it
	VITE_BASE_URL=$(VITE_FRONTEND_BASE_URL_) \
	VITE_DB_URL=$(VITE_FRONTEND_BASE_URL_)/data/db.sqlite3 \
	VITE_BLOBS_URL=$(VITE_FRONTEND_BASE_URL_)/data/blobs \
		npm run build

deploy: build
	test -n "$(FRONTEND_DEPLOY_TARGET_SSH_HOST_)" # missing env var, please set up env file from env.example and source it
	rm -rf dist/_data
	chmod 755 dist/
	rsync --verbose --archive --compress --rsh=ssh dist/ $(FRONTEND_DEPLOY_TARGET_SSH_HOST_)

DOCKER_FLAGS = $(DOCKER_CLI_FLAGS)
DOCKER_FLAGS += --build-arg DOCKER_BASE_IMAGE="$(DOCKER_BASE_IMAGE)"

docker_build:
	docker buildx build $(DOCKER_FLAGS)     \
		--target=export              \
		--output=dist                \
		.

list:
	@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'
