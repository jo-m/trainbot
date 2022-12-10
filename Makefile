.PHONY: format lint test bench check build_host build_docker clean

DOCKER_BUILDER_IMG_TAG = trainbot-builder
DOCKER_TMP_CONTAINER_NAME = trainbot-tmp-container

DOCKER_BASE_IMAGE = ubuntu:jammy-20221130
GO_VERSION = 1.19.4
GO_ARCHIVE_SHA256 = c9c08f783325c4cf840a94333159cc937f05f75d36a8b307951d5bd959cf2ab8
GO_STATICCHECK_VERSION = 2022.1.3

DEFAULT: format build_host

format:
	cd pkg/pmatch && clang-format -i -style=Google *.h *.c
	gofmt -w .
	go mod tidy

lint:
	gofmt -l .; test -z "$$(gofmt -l .)"
	go run honnef.co/go/tools/cmd/staticcheck@$(GO_STATICCHECK_VERSION) ./...
	go vet ./...

test:
	go test -race -v ./...

bench:
	go test -v -bench=. ./...

check: lint test bench

build_host:
	mkdir -p out
	go build -ldflags "-linkmode external -extldflags -static" -o out/trainbot ./cmd/trainbot
	go build -ldflags "-linkmode external -extldflags -static" -o out/confighelper ./cmd/confighelper
	go build -ldflags "-linkmode external -extldflags -static" -o out/pmatch ./examples/pmatch

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
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/trainbot-arm6 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/confighelper-arm6 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/pmatch-arm6 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/trainbot-arm64 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/confighelper-arm64 out/
	docker cp $(DOCKER_TMP_CONTAINER_NAME):/out/pmatch-arm64 out/

	# Remove temporary container
	docker rm -f $(DOCKER_TMP_CONTAINER_NAME)

clean:
	rm -rf out/
	rm -f prof-*.gz
	rm -rf imgs/
