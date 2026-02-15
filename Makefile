VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)
BINARY := bin/tuner

.PHONY: build clean test lint run

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tuner

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	golangci-lint run ./...

run: build
	./$(BINARY)

release:
	goreleaser release --clean
