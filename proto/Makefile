.PHONY: all build clean

PROTOC := /usr/bin/protoc

all: clean build

build:
	$(PROTOC) -I ./ --go_out=./ --go_opt=module=hermes ./*.proto

clean:
	$(RM) -f ./*.pb.go
