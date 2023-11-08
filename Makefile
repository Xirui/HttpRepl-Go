SHELL := /bin/bash

BINARY_NAME = httpreplgo

install:
	@go mod tidy
	@echo "All dependencies are installed."

build:
	@go build -ldflags "-s -w" -o $(BINARY_NAME)

run:
	go run . -h

clean:
	@rm -rf $(BINARY_NAME)
	@go clean
	@echo "All build and temporary files have been removed."

.PHONY: install build run clean