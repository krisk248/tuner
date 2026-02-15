VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)
BINARY := bin/tuner

.PHONY: build build-all clean test lint vet run release

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tuner

build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/tuner-linux-amd64 ./cmd/tuner
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/tuner-linux-arm64 ./cmd/tuner

clean:
	rm -rf bin/ dist/

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

check: vet test

run: build
	./$(BINARY)

release:
	goreleaser release --clean

release-snapshot:
	goreleaser release --snapshot --clean
