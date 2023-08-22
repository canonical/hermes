.PHONY: all build clean install

RM := rm -f
RMDIR := rm -rf
MKDIRP := mkdir -p
CPDIR := cp -r
SRC_FRONTEND_DIR := ./build/
DST_FRONTEND_DIR := $(if $(DESTDIR),$(DESTDIR),$(HOME))/hermes/frontend/

all: clean build

build:
	npm install
	protoc -I ../proto --js_out=import_style=commonjs,binary:. ../proto/*.proto
	npm run build

clean:
	$(RMDIR) ./build
	$(RM) ./*pb.js
	$(RMDIR) ./node_modules

install:
	$(MKDIRP) $(DST_FRONTEND_DIR)
	$(CPDIR) $(SRC_FRONTEND_DIR)* $(DST_FRONTEND_DIR)
