.PHONY: format lint test bench check build_host build_docker clean run_confighelper run_camera run_rec run_videofile

DOCKER_BUILDER_IMG_TAG = trainbot-builder
DOCKER_TMP_CONTAINER_NAME = trainbot-tmp-container

DOCKER_BASE_IMAGE = debian:bullseye-20230320
GO_VERSION = 1.20.2
GO_ARCHIVE_SHA256 = 4eaea32f59cde4dc635fbc42161031d13e1c780b87097f4b4234cfce671f1768
GO_STATICCHECK_VERSION = 2023.1.3

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
	mkdir -p build
	go build -o build/trainbot ./cmd/trainbot
	go build -o build/confighelper ./cmd/confighelper
	go build -o build/pmatch ./examples/pmatch

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
	go build -o trainbot ./cmd/trainbot/
	./trainbot \
		--log-pretty \
		--log-level=info \
		\
		--input="vids/phone/day.mp4" \
		-X 800 -Y 450 -W 300 -H 300

# Build and copy to Raspberry Pi, outside docker
deploy_confighelper:
	CGO_ENABLED=1              \
    CC=aarch64-linux-gnu-gcc   \
    GOOS=linux                 \
    GOARCH=arm64               \
    GOARM=7 \
		go build -o out/confighelper-arm64 ./cmd/confighelper
	scp out/confighelper-arm64 pi4:

# Build and copy to Raspberry Pi, outside docker
deploy_trainbot:
	CGO_ENABLED=1              \
    CC=aarch64-linux-gnu-gcc   \
    GOOS=linux                 \
    GOARCH=arm64               \
    GOARM=7 \
		go build -o out/trainbot-arm64 ./cmd/trainbot
	scp out/trainbot-arm64 pi4:
