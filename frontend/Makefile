.PHONY: all build clean install

RM               := /usr/bin/rm
NPM              := /usr/bin/npm
MKDIR            := /usr/bin/mkdir
CP               := /usr/bin/cp
PROTOC           := /usr/bin/protoc
SRC_FRONTEND_DIR := ./build/
DST_FRONTEND_DIR := $(HOME)/frontend/

all: clean build

build:
	$(NPM) install
	$(PROTOC) -I ../proto --js_out=import_style=commonjs,binary:. ../proto/*.proto
	$(NPM) run build

clean:
	$(RM) -rf ./build
	$(RM) -f ./*pb.js

install:
	$(MKDIR) -p $(DST_FRONTEND_DIR)
	$(CP) -r $(SRC_FRONTEND_DIR)* $(DST_FRONTEND_DIR)
