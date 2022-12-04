.PHONY: test bench

test:
	go test -count 1 -race -v ./...

bench:
	go test -v -bench=. ./...
