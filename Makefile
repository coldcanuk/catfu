.PHONY: build test vet lint staticcheck tidy install clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/coldcanuk/catfu/pkg/version.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/catfu ./cmd/catfu

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/catfu

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

staticcheck:
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed; skip"

lint: vet staticcheck

clean:
	rm -rf bin/
