.PHONY: all build run test

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod

# Binary names
BINARY=bin/portx

# Directories
CMD_APP=cmd/portx

all: init build test

init:
	go mod tidy

run:
	$(GOCMD) run ./$(CMD_APP)

fmt:
	$(GOFMT) -w ./...

vet:
	$(GOVET) ./...

lint: fmt vet

clean:
	rm -rf $(BINARY)

build:
	$(GOBUILD) -o $(BINARY) ./$(CMD_APP)

test:
	$(GOTEST) -v -race ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

deps:
	$(GOMOD) download
	$(GOMOD) tidy
