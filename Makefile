.PHONY: all check auto_install generate build install install_bin install_ui clean

RM := rm -f
RMDIR := rm -rf
MKDIRP := mkdir -p
CPDIR := cp -r
PROTO_DIR := proto
FRONTEND_DIR := frontend
GRAFANA_DIR := grafana_app
INSTALL_DIR := ./install/
METADATA_DIR := $(if $(DESTDIR),$(DESTDIR),$(HOME))/hermes/
SRC_CONFIG_DIR := ./config/
DST_CONFIG_DIR := $(METADATA_DIR)/config/
INSTALL_BIN := install -m 755
DST_BIN_DIR := /usr/sbin/
CFLAGS := -O2 -g -Wall -Werror $(CFLAGS)

all: clean build

check:
	go vet ./...

auto_install:
ifeq ($(shell dpkg -s curl 2> /dev/null; echo $$?), 1)
	apt install -y curl
endif
ifeq ($(shell dpkg -s llvm 2> /dev/null; echo $$?), 1)
	apt install -y llvm
endif
ifeq ($(shell dpkg -s clang 2> /dev/null; echo $$?), 1)
	apt install -y clang
endif
ifeq ($(shell dpkg -s libbpf-dev 2> /dev/null; echo $$?), 1)
	apt install -y libbpf-dev
endif
ifeq ($(shell dpkg -s gcc-multilib 2> /dev/null; echo $$?), 1)
	apt install -y gcc-multilib
endif
ifeq ($(shell dpkg -s protobuf-compiler 2> /dev/null; echo $$?), 1)
	apt install -y protobuf-compiler
endif
ifeq ($(shell dpkg -s pkg-config 2> /dev/null; echo $$?), 1)
	apt install -y pkg-config
endif
ifeq ($(shell dpkg -s nodejs 2> /dev/null; echo $$?), 1)
	apt install -y ca-certificates gnupg
	mkdir -p /etc/apt/keyrings
	curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg
	echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_18.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list
	apt update
	apt install -y nodejs
endif
	snap install go --classic
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

build: auto_install ui backend grafana

generate: export BPF_CLANG := clang
generate: export BPF_CFLAGS := $(CFLAGS)
generate: auto_install
	go generate ./backend/ebpf/...

backend: generate
	make -C $(PROTO_DIR) build
	go build -ldflags "-X main.metadataDir=$(METADATA_DIR)" -o $(INSTALL_DIR) ./...

ui: auto_install
	make -C $(FRONTEND_DIR) build

grafana: auto_install
	make -C $(GRAFANA_DIR) build

install: install_bin install_ui

install_bin:
	$(MKDIRP) $(DESTDIR)$(DST_BIN_DIR)
	$(INSTALL_BIN) $(INSTALL_DIR)* $(DESTDIR)$(DST_BIN_DIR)
	$(MKDIRP) $(DST_CONFIG_DIR)
	$(CPDIR) $(SRC_CONFIG_DIR)* $(DST_CONFIG_DIR)

install_ui:
	make -C $(FRONTEND_DIR) install

clean:
ifneq ($(shell which go),)
	go clean
endif
	$(RMDIR) $(INSTALL_DIR)
	$(RM) ./backend/ebpf/*/bpf_bpfeb*.go
	$(RM) ./backend/ebpf/*/bpf_bpfeb*.o
	$(RM) ./backend/ebpf/*/bpf_bpfel*.go
	$(RM) ./backend/ebpf/*/bpf_bpfel*.o
	make -C $(PROTO_DIR) clean
	make -C $(FRONTEND_DIR) clean
	make -C $(GRAFANA_DIR) clean
