.PHONY: all check auto_install generate build install install_bin install_ui clean

GO             := go
RM             := rm -f
RMDIR          := rm -rf
MKDIRP         := mkdir -p
CPDIR          := cp -r
CLANG          := clang
MAKE           := make
SNAP           := snap
APT            := apt
DPKG           := dpkg
ECHO           := echo
WHICH          := which
CURL           := curl
PROTO_DIR      := proto
FRONTEND_DIR   := frontend
BUILD_DIR      := ./build/
SRC_CONFIG_DIR := ./config/
DST_CONFIG_DIR := $(HOME)/config/
INSTALL_BIN    := install -m 755
DST_BIN_DIR    := /usr/sbin/
CFLAGS         := -O2 -g -Wall -Werror $(CFLAGS)

all: clean build

check:
	$(GO) vet ./...

auto_install:
ifeq ($(shell $(DPKG) -s curl 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y curl
endif
ifeq ($(shell $(DPKG) -s llvm 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y llvm
endif
ifeq ($(shell $(DPKG) -s clang 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y clang
endif
ifeq ($(shell $(DPKG) -s libbpf-dev 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y libbpf-dev
endif
ifeq ($(shell $(DPKG) -s gcc-multilib 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y gcc-multilib
endif
ifeq ($(shell $(DPKG) -s protobuf-compiler 2> /dev/null; $(ECHO) $$?), 1)
	$(APT) install -y protobuf-compiler
endif
	$(CURL) -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && apt install -y nodejs
	$(SNAP) install go --classic
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go

generate: export BPF_CLANG := $(CLANG)
generate: export BPF_CFLAGS := $(CFLAGS)
generate:
	$(GO) generate ./backend/ebpf/...

build: auto_install generate
	$(MAKE) -C $(PROTO_DIR) build
	$(GO) build -o $(BUILD_DIR) ./...
	$(MAKE) -C $(FRONTEND_DIR) build

install: install_bin install_ui

install_bin:
	$(INSTALL_BIN) $(BUILD_DIR)* $(DST_BIN_DIR)
	$(MKDIRP) $(DST_CONFIG_DIR)
	$(CPDIR) $(SRC_CONFIG_DIR)* $(DST_CONFIG_DIR)

install_ui:
	$(MAKE) -C $(FRONTEND_DIR) install

clean:
ifneq ($(shell $(WHICH) $(GO)),)
	$(GO) clean
endif
	$(RMDIR) $(BUILD_DIR)
	$(RM) ./backend/ebpf/*/bpf_bpfeb*.go
	$(RM) ./backend/ebpf/*/bpf_bpfeb*.o
	$(RM) ./backend/ebpf/*/bpf_bpfel*.go
	$(RM) ./backend/ebpf/*/bpf_bpfel*.o
	$(MAKE) -C $(PROTO_DIR) clean
	$(MAKE) -C $(FRONTEND_DIR) clean
