SHELL := /bin/bash

.PHONY: build clean

build:
	@echo "Building sshmenu (CGO disabled, pure-Go)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags "-s -w" -o sshmenu ./...

clean:
	@rm -f sshmenu
	@echo "Cleaned."
