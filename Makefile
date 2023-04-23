.PHONY: format lint test test_more bench check build_host build_arm64 build_docker clean run_confighelper run_camera run_videofile

DOCKER_BUILDER_IMG_TAG = trainbot-builder
DOCKER_TMP_CONTAINER_NAME = trainbot-tmp-container

DOCKER_BASE_IMAGE = debian:bullseye-20230320
GO_VERSION = 1.20.3
GO_ARCHIVE_SHA256 = 979694c2c25c735755bf26f4f45e19e64e4811d661dd07b8c010f7a8e18adfca
GO_STATICCHECK_VERSION = 2023.1.3

TRAINBOT_DEPLOY_TARGET_SSH_HOST_ = ${TRAINBOT_DEPLOY_TARGET_SSH_HOST}

DEFAULT: format build_host build_arm64

format:
	cd pkg/pmatch && clang-format -i -style=Google *.h *.c
	gofmt -w .
	go mod tidy

lint:
	gofmt -l .; test -z "$$(gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@$(GO_STATICCHECK_VERSION) -checks=all ./...
	go run github.com/mgechev/revive@latest -set_exit_status ./...
	go run github.com/securego/gosec/v2/cmd/gosec@latest -exclude G307 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

test:
	go test -race -v ./...

test_more:
	# This needs additional test data which is not committed to the repo.
	# Instructions:
	#	curl -o internal/pkg/stitch/testdata/more-testdata.zip https://trains.jo-m.ch/testdata.zip
	#	unzip -d internal/pkg/stitch/testdata internal/pkg/stitch/testdata/more-testdata.zip
	go test -v --tags=moretests -run Test_AutoStitcher_Set ./...

bench:
	go test -v -run= -bench=. ./...

check: lint test bench

build_host: export CGO_ENABLED=1
build_host: export CC=gcc
build_host:
	mkdir -p build
	go build -o build/trainbot ./cmd/trainbot
	go build -o build/confighelper ./cmd/confighelper
	go build -o build/pmatch ./examples/pmatch

build_arm64: export CGO_ENABLED=1
build_arm64: export CC=aarch64-linux-gnu-gcc
build_arm64: export GOOS=linux
build_arm64: export GOARCH=arm64
build_arm64: export GOARM=7
build_arm64:
	mkdir -p build
	go build -o build/trainbot-arm64 ./cmd/trainbot
	go build -o build/confighelper-arm64 ./cmd/confighelper
	go build -o build/pmatch-arm64 ./examples/pmatch

build_docker:
	# Build
	docker build \
		--tag "$(DOCKER_BUILDER_IMG_TAG)"                                 \
		--build-arg DOCKER_BASE_IMAGE="$(DOCKER_BASE_IMAGE)"              \
		--build-arg GO_VERSION="$(GO_VERSION)"                            \
		--build-arg GO_ARCHIVE_SHA256="$(GO_ARCHIVE_SHA256)"              \
		--build-arg GO_STATICCHECK_VERSION="$(GO_STATICCHECK_VERSION)"    \
		.

	# Start temporary container
	mkdir -p build
	docker rm -f $(DOCKER_TMP_CONTAINER_NAME) || true
	docker create -ti --name $(DOCKER_TMP_CONTAINER_NAME) $(DOCKER_BUILDER_IMG_TAG)

	# Copy
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/trainbot build/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/confighelper build/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/pmatch build/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/trainbot-arm64 build/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/confighelper-arm64 build/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/build/pmatch-arm64 build/

	# Remove temporary container
	docker rm -f $(DOCKER_TMP_CONTAINER_NAME)

clean:
	rm -rf build/
	rm -f prof-*.gz

run_confighelper:
	go run ./cmd/confighelper/ --input /dev/video2 --live-reload

run_camera:
	go run ./cmd/trainbot \
		--log-pretty \
		--input /dev/video2 --camera-format-fourcc MJPG --camera-w 1920 --camera-h 1080 \
		-X 1064 -Y 178 -W 366 -H 334

run_videofile:
	go build -o build/trainbot ./cmd/trainbot/
	./build/trainbot \
		--log-pretty \
		--log-level=info \
		\
		--input="vids/day.mp4" \
		-X 800 -Y 450 -W 300 -H 300

deploy_trainbot: build_arm64
	test -n "$(TRAINBOT_DEPLOY_TARGET_SSH_HOST_)" # missing env var
	scp env $(TRAINBOT_DEPLOY_TARGET_SSH_HOST_):
	scp build/trainbot-arm64 $(TRAINBOT_DEPLOY_TARGET_SSH_HOST_):

deploy_confighelper: build_arm64
	test -n "$(TRAINBOT_DEPLOY_TARGET_SSH_HOST_)" # missing env var
	scp build/confighelper-arm64 $(TRAINBOT_DEPLOY_TARGET_SSH_HOST_):
