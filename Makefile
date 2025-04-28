.PHONY: build install clean test

# Binary name
BINARY_NAME=anytype-cli

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Build directory
BUILD_DIR=bin

# Output binary
OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)

# Default target
all: build

# Init project 
init:
	mkdir -p $(BUILD_DIR)
	$(GOMOD) tidy

# Build binary
build: init
	$(GOBUILD) -o $(OUTPUT) -v

# Install binary
install: build
	install -m755 $(OUTPUT) /usr/local/bin/$(BINARY_NAME)

# Run the application
run:
	$(GORUN) main.go

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Build binaries for multiple platforms
release: init
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -v
	# MacOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v
