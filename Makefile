.PHONY: all check build clean

GO             := /usr/bin/go
RM             := /usr/bin/rm
MKDIR          := /usr/bin/mkdir
CP             := /usr/bin/cp
INSTALL        := /usr/bin/install
CLANG          := /usr/bin/clang-15
STRIP          := /usr/bin/llvm-strip-15
OUTPUT_DIR     := ./output/
SRC_CONFIG_DIR := ./config/
DST_CONFIG_DIR := /root/config/
INSTALL_BIN    := /usr/bin/install -m 755
DST_BIN_DIR    := /usr/sbin/
CFLAGS         := -O2 -g -Wall -Werror $(CFLAGS)

all: clean build

check:
	$(GO) vet ./...

generate: export BPF_CLANG := $(CLANG)
generate: export BPF_CFLAGS := $(CFLAGS)
generate:
	$(GO) generate ./backend/ebpf/...

build: generate
	$(GO) build -o $(OUTPUT_DIR) ./...

install:
	$(MKDIR) -p $(DST_CONFIG_DIR)
	$(CP) -r $(SRC_CONFIG_DIR)* $(DST_CONFIG_DIR)

	$(INSTALL_BIN) $(OUTPUT_DIR)* $(DST_BIN_DIR)

clean:
	$(GO) clean
	$(RM) -rf $(OUTPUT_DIR)
	$(RM) -f ./backend/ebpf/*/bpf_bpfeb.go
	$(RM) -f ./backend/ebpf/*/bpf_bpfeb.o
	$(RM) -f ./backend/ebpf/*/bpf_bpfel.go
	$(RM) -f ./backend/ebpf/*/bpf_bpfel.o
