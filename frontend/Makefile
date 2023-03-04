.PHONY: all build clean install

RM               := rm -f
RMDIR            := rm -rf
NPM              := npm
MKDIRP           := mkdir -p
CPDIR            := cp -r
PROTOC           := protoc
SRC_FRONTEND_DIR := ./build/
DST_FRONTEND_DIR := $(HOME)/frontend/

all: clean build

build:
	$(NPM) install
	$(PROTOC) -I ../proto --js_out=import_style=commonjs,binary:. ../proto/*.proto
	$(NPM) run build

clean:
	$(RMDIR) ./build
	$(RM) ./*pb.js
	$(RMDIR) ./node_modules

install:
	$(MKDIRP) $(DST_FRONTEND_DIR)
	$(CPDIR) $(SRC_FRONTEND_DIR)* $(DST_FRONTEND_DIR)
