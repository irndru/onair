export CGO_ENABLED = 0

.PHONY: build install test lint

build:
	go build -o videobg ./cmd/videobg

install:
	go install ./cmd/videobg

test:
	go test ./...

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...
	GOOS=linux go vet ./...
