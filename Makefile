.PHONY: format lint test test_vk test_more bench bench_vk check build_host build_host_vk build_arm64 docker_build docker_lint docker_test docker_test_more docker_bench clean run_confighelper run_camera run_videofile list

# https://hub.docker.com/_/debian
DOCKER_BASE_IMAGE = debian:bullseye-20240926
# https://go.dev/dl/
GO_VERSION = 1.22.6
GO_ARCHIVE_SHA256 = 999805bed7d9039ec3da1a53bfbcafc13e367da52aa823cb60b68ba22d44c616
# https://github.com/dominikh/go-tools/releases
GO_STATICCHECK_VERSION = 2024.1.1
# https://github.com/mgechev/revive/releases
GO_REVIVE_VERSION = v1.4.0
# https://github.com/securego/gosec/releases
GO_SEC_VERSION = v2.21.4
# https://pkg.go.dev/golang.org/x/vuln?tab=versions
GO_VULNCHECK_VERSION = v1.1.3

DEFAULT: format build_host build_arm64

GO_BUILD_TAGS =

format:
	bash -c "shopt -s globstar; clang-format -i **/*.c **/*.h **/*.comp"
	gofmt -w .
	go mod tidy

lint:
	bash -c "shopt -s globstar; clang-format --dry-run --Werror **/*.c **/*.h **/*.comp"
	gofmt -l .; test -z "$$(gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@$(GO_STATICCHECK_VERSION) -checks=all ./...
	go run github.com/mgechev/revive@$(GO_REVIVE_VERSION) -set_exit_status ./...
	# G307: Poor file permissions used when creating a file with os.Create
	# G115: Type conversion which leads to integer overflow
	go run github.com/securego/gosec/v2/cmd/gosec@$(GO_SEC_VERSION) -exclude G307,G115 ./...
	go run golang.org/x/vuln/cmd/govulncheck@$(GO_VULNCHECK_VERSION) ./...

generate:
	go generate --tags=$(GO_BUILD_TAGS) ./...

test: generate
	go test -v --tags=$(GO_BUILD_TAGS) ./...

test_vk: GO_BUILD_TAGS = vk
test_vk: test

test_more: GO_BUILD_TAGS = moretests
test_more:
	# This needs additional test data which is not committed to the repo.
	# Instructions:
	#	curl -o internal/pkg/stitch/testdata/more-testdata.zip https://trains.jo-m.ch/testdata.zip
	#	unzip -d internal/pkg/stitch/testdata internal/pkg/stitch/testdata/more-testdata.zip
	go test -v --tags=$(GO_BUILD_TAGS) -timeout=30m -run Test_AutoStitcher_Set ./...

bench:
	go test -v --tags=$(GO_BUILD_TAGS) -run=Nothing -bench=Benchmark_ ./...

bench_vk: GO_BUILD_TAGS = vk
bench_vk: bench

check: lint test bench

build_host: export CGO_ENABLED=1
build_host: export CC=gcc
build_host:
	mkdir -p build
	go build --tags=$(GO_BUILD_TAGS) -o build/trainbot ./cmd/trainbot
	go build --tags=$(GO_BUILD_TAGS) -o build/confighelper ./cmd/confighelper
	go build --tags=$(GO_BUILD_TAGS) -o build/pmatch ./examples/pmatch

build_host_vk: GO_BUILD_TAGS = vk
build_host_vk: generate build_host
build_host_vk:
	go build --tags=$(GO_BUILD_TAGS) -o build/pmatchVk ./examples/pmatchVk

TRAINBOT_AARCH_CROSS ?= aarch64-linux-gnu-gcc
build_arm64: export CGO_ENABLED=1
build_arm64: export CC=$(TRAINBOT_AARCH_CROSS)
build_arm64: export GOOS=linux
build_arm64: export GOARCH=arm64
build_arm64: export GOARM=7
build_arm64:
	mkdir -p build
	go build --tags=$(GO_BUILD_TAGS) -o build/trainbot-arm64 ./cmd/trainbot
	go build --tags=$(GO_BUILD_TAGS) -o build/confighelper-arm64 ./cmd/confighelper
	go build --tags=$(GO_BUILD_TAGS) -o build/pmatch-arm64 ./examples/pmatch

DOCKER_FLAGS = $(DOCKER_CLI_FLAGS)
DOCKER_FLAGS += --build-arg DOCKER_BASE_IMAGE="$(DOCKER_BASE_IMAGE)"
DOCKER_FLAGS += --build-arg GO_VERSION="$(GO_VERSION)"
DOCKER_FLAGS += --build-arg GO_ARCHIVE_SHA256="$(GO_ARCHIVE_SHA256)"
DOCKER_FLAGS += --build-arg GO_STATICCHECK_VERSION="$(GO_STATICCHECK_VERSION)"

docker_image:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=source                   \
		--load                            \
		--tag trainbot-source:latest      \
		.

docker_build:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=export                   \
		--output=build                    \
		.

docker_lint:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=lint                     \
		.

docker_test:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=test                     \
		.

docker_test_more:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=test_more                \
		.

docker_bench:
	docker buildx build $(DOCKER_FLAGS)   \
		--target=bench                    \
		.

clean:
	rm -rf build/
	rm -f prof-*.gz
	find . -name '*.spv' -delete

run_confighelper:
	go run ./cmd/confighelper/ --input /dev/video2 --live-reload

run_camera:
	go run ./cmd/trainbot \
		--log-pretty \
		\
		--enable-upload=false \
		--input /dev/video2 \
		--camera-format-fourcc MJPG \
		--camera-w 1920 --camera-h 1080 \
		-X 1064 -Y 178 -W 366 -H 334

run_videofile:
	go build --tags=$(GO_BUILD_TAGS) -o build/trainbot ./cmd/trainbot/
	./build/trainbot \
		--log-pretty \
		--log-level=info \
		\
		--enable-upload=false \
		--input="internal/pkg/stitch/testdata/set0/day.mp4" \
		-X 0 -Y 0 -W 300 -H 300

# Usage: make deploy_trainbot host=$TRAINBOT_DEPLOY_TARGET_SSH_HOST
# Example: make deploy_trainbot host=pi@10.20.0.12
deploy_trainbot: docker_build
	test -n "$(host)" # missing target host, usage: make deploy_trainbot host=TRAINBOT_DEPLOY_TARGET_SSH_HOST !

	ssh $(host) mkdir -p trainbot/
	scp env $(host):trainbot/
	ssh $(host) systemctl --user stop trainbot.service
	scp build/trainbot-arm64 $(host):trainbot/

	ssh $(host) mkdir -p .config/systemd/user/
	scp trainbot.service $(host):.config/systemd/user/

	ssh $(host) loginctl enable-linger
	ssh $(host) systemctl --user enable trainbot.service
	ssh $(host) systemctl --user start trainbot.service

# Usage: make deploy_confighelper host=TRAINBOT_DEPLOY_TARGET_SSH_HOST
# Example: make deploy_confighelper host=pi@10.20.0.12
deploy_confighelper: docker_build
	test -n "$(host)" # missing target host, usage: make deploy_confighelper host=TRAINBOT_DEPLOY_TARGET_SSH_HOST !
	ssh $(host) mkdir -p trainbot/
	scp build/confighelper-arm64 $(host):trainbot/

list:
	@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'
