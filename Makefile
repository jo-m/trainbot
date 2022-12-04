.PHONY: test bench compare

test:
	go test -count 1 -race -v ./...

bench:
	go test -v -bench=. ./...

compare:
	go test -v -bench='SearchGray|SearchGrayOpt' ./...
