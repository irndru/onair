export CGO_ENABLED = 0

.PHONY: build install test lint fmt

build:
	go build -o onair ./cmd/onair

install:
	go install ./cmd/onair

test:
	go test ./...

lint:
	golangci-lint run
	GOOS=linux golangci-lint run

fmt:
	golangci-lint fmt
