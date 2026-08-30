.PHONY: all build install run clean test lint release

BINARY   := anytype-cli
BUILD_DIR := bin
MODULE   := github.com/epheo/anytype-cli
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/cmd.version=$(VERSION)

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .

install: build
	install -m755 $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

run:
	go run -ldflags "$(LDFLAGS)" . $(ARGS)

test:
	go test ./...

lint:
	gofmt -l .
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

# Cross-compile release binaries.
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
release:
	mkdir -p $(BUILD_DIR)
	$(foreach p,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(p))) GOARCH=$(word 2,$(subst /, ,$(p))) \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-$(subst /,-,$(p))$(if $(findstring windows,$(p)),.exe,) . ;)
