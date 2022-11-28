.PHONY: all check build clean

GO             := /usr/bin/go
RM             := /usr/bin/rm
MKDIR          := /usr/bin/mkdir
CP             := /usr/bin/cp
INSTALL        := /usr/bin/install
OUTPUT_DIR     := ./output/
SRC_CONFIG_DIR := ./config/
DST_CONFIG_DIR := /root/config/
INSTALL_BIN    := /usr/bin/install -m 755
DST_BIN_DIR    := /usr/sbin/

all: clean build

check:
	$(GO) vet ./...

build:
	$(GO) build -o $(OUTPUT_DIR) ./...

install:
	$(MKDIR) -p $(DST_CONFIG_DIR)
	$(CP) -r $(SRC_CONFIG_DIR)* $(DST_CONFIG_DIR)

	$(INSTALL_BIN) $(OUTPUT_DIR)* $(DST_BIN_DIR)

clean:
	$(GO) clean
	$(RM) -rf $(OUTPUT_DIR)
