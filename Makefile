.PHONY: format lint test test_more bench check build_host build_arm64 docker_build docker_lint docker_test docker_test_more docker_bench clean run_confighelper run_camera run_videofile list

# https://hub.docker.com/_/debian
DOCKER_BASE_IMAGE = debian:bullseye-20231120
# https://go.dev/dl/
GO_VERSION = 1.21.4
GO_ARCHIVE_SHA256 = 73cac0215254d0c7d1241fa40837851f3b9a8a742d0b54714cbdfb3feaf8f0af
# https://github.com/dominikh/go-tools/releases
GO_STATICCHECK_VERSION = 2023.1.6

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
	go test -v --tags=moretests -timeout=30m -run Test_AutoStitcher_Set ./...

bench:
	go test -v -run=Nothing -bench=Benchmark_ ./...

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

DOCKER_FLAGS = $(DOCKER_CLI_FLAGS)
DOCKER_FLAGS += --build-arg DOCKER_BASE_IMAGE="$(DOCKER_BASE_IMAGE)"
DOCKER_FLAGS += --build-arg GO_VERSION="$(GO_VERSION)"
DOCKER_FLAGS += --build-arg GO_ARCHIVE_SHA256="$(GO_ARCHIVE_SHA256)"
DOCKER_FLAGS += --build-arg GO_STATICCHECK_VERSION="$(GO_STATICCHECK_VERSION)"

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
	go build -o build/trainbot ./cmd/trainbot/
	./build/trainbot \
		--log-pretty \
		--log-level=info \
		\
		--enable-upload=false \
		--input="internal/pkg/stitch/testdata/set0/day.mp4" \
		-X 0 -Y 0 -W 300 -H 300

# Usage: make deploy_trainbot host=TRAINBOT_DEPLOY_TARGET_SSH_HOST
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
