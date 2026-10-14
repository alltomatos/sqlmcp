.PHONY: all build test clean run

BINARY_NAME=sqlmcp.exe

all: test build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/sqlmcp

test:
	go test -v ./...

clean:
	go clean
	if exist bin rmdir /s /q bin

run:
	go run ./cmd/sqlmcp
