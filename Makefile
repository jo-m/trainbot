.PHONY: format lint test bench check build_host build_docker clean run_confighelper run_camera run_rec

DOCKER_BUILDER_IMG_TAG = trainbot-builder
DOCKER_TMP_CONTAINER_NAME = trainbot-tmp-container

DOCKER_BASE_IMAGE = debian:bullseye-20230208
GO_VERSION = 1.20.1
GO_ARCHIVE_SHA256 = 000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02
GO_STATICCHECK_VERSION = 2023.1.2

DEFAULT: format build_host

format:
	cd pkg/pmatch && clang-format -i -style=Google *.h *.c
	gofmt -w .
	go mod tidy

lint:
	gofmt -l .; test -z "$$(gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@$(GO_STATICCHECK_VERSION) -checks=all ./...
	go run github.com/mgechev/revive@latest -set_exit_status ./...
	go run github.com/securego/gosec/v2/cmd/gosec@latest ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

test:
	go test -race -v ./...

bench:
	go test -v -run= -bench=. ./...

check: lint test bench

build_host:
	mkdir -p out
	go build -o out/trainbot ./cmd/trainbot
	go build -o out/confighelper ./cmd/confighelper
	go build -o out/pmatch ./examples/pmatch

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
	mkdir -p out
	docker rm -f $(DOCKER_TMP_CONTAINER_NAME) || true
	docker create -ti --name $(DOCKER_TMP_CONTAINER_NAME) $(DOCKER_BUILDER_IMG_TAG)

	# Copy
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/trainbot out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/confighelper out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/pmatch out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/trainbot-arm64 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/confighelper-arm64 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/pmatch-arm64 out/

	# Remove temporary container
	docker rm -f $(DOCKER_TMP_CONTAINER_NAME)

clean:
	rm -rf out/
	rm -f prof-*.gz
	rm -rf imgs/

run_confighelper:
	go run ./cmd/confighelper/ --camera-device /dev/video2

run_camera:
	go run ./cmd/trainbot \
		--log-pretty \
		--camera-device /dev/video2 --camera-format-fourcc MJPG --camera-w 1920 --camera-h 1080 \
		-X 1064 -Y 178 -W 366 -H 334

run_rec:
	# go tool pprof trainbot prof-cpu.gz
	# go tool pprof trainbot prof-heap-XX.gz
	go build -o trainbot ./cmd/trainbot/
	./trainbot \
		--log-pretty \
		--log-level=debug \
		--cpu-profile \
		--heap-profile \
		\
		--video-file="imgs/20221208_092919.709_+01:00" \
		-X 0 -Y 0 -W 300 -H 350
