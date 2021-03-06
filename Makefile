NAME       := git-lfs-cloud
OUTPUT_DIR := build
TARGET     := $(OUTPUT_DIR)/$(NAME)

SRCS    := $(wildcard *.go)
GOBUILD := go build -v
GOOS     = $(shell go env GOOS)
GOARCH   = $(shell go env GOARCH)

.PHONY: build
build: $(TARGET) \

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf build

$(TARGET): $(SRCS)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o $@ $^

.PHONY: run
run: $(SRCS)
	go run $^ ./test_config.toml