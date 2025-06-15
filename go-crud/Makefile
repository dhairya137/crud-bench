.PHONY: build run clean test

# Binary name
BINARY_NAME=crud-bench

# Build directory
BUILD_DIR=./bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

all: clean build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/crud-bench

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) tidy

# Example targets for common benchmarks
mysql-test: build
	$(BUILD_DIR)/$(BINARY_NAME) -d mysql -i mysql:8.1 -s 1000 -c 4 -t 8 -a '[{"name":"count_all","samples":3,"projection":"COUNT"}]'

postgres-test: build
	$(BUILD_DIR)/$(BINARY_NAME) -d postgres -i postgres:14.2 -s 1000 -c 4 -t 8 -a '[{"name":"count_all","samples":3,"projection":"COUNT"}]'