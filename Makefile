.PHONY: test bench compare run_videofile run_camera format lint check clean

test:
	go test -count 1 -race -v ./...

bench:
	go test -v -bench=. ./...

compare:
	go test -v -bench='SearchGray|SearchGrayOpt' ./...

run_videofile:
	# go tool pprof trainbot prof.gz
	go build -o trainbot ./cmd/trainbot/
	./trainbot \
		--log-pretty \
		-X 800 -Y 450 -W 300 -H 300 \
		--cpu-profile prof.gz \
		--video-file="vids/phone/VID_20220626_104921284-00.00.06.638-00.00.14.810.mp4"

run_camera:
	go build -o trainbot ./cmd/trainbot/
	./trainbot \
		--log-pretty \
		-X 800 -Y 590 -W 300 -H 350 \
		--camera-device /dev/video2 \
		--format-fourcc MJPG \
		--framesz-w 1920 \
		--framesz-h 1080

format:
	cd pkg/pmatch && clang-format -i -style=Google *.h *.c
	gofmt -w .
	go mod tidy

lint:
	gofmt -l .; test -z "$$(gofmt -l .)"
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	go vet ./...

check: lint test

clean:
	rm -f trainbot
	rm -f prof.gz
	rm -rf imgs
	mkdir -p imgs
