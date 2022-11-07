.PHONY: all check build clean

GO         ?= go
RM          = rm
OUTPUT_DIR  = ./output/

all: clean build

check:
	$(GO) vet ./...

build:
	$(GO) build -o $(OUTPUT_DIR) ./...

clean:
	$(GO) clean
	$(RM) -rf $(OUTPUT_DIR)
