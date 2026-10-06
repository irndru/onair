export CGO_ENABLED = 0

.PHONY: build install test lint

build:
	go build -o onair ./cmd/onair

install:
	go install ./cmd/onair

test:
	go test ./...

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...
	GOOS=linux go vet ./...
