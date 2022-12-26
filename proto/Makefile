.PHONY: all build clean

GO     := go
PROTOC := /usr/bin/protoc

all: clean build

build:
	PATH=$(PATH):$(shell $(GO) env GOPATH)/bin $(PROTOC) -I ./ --go_out=./ --go_opt=module=hermes ./*.proto

clean:
	$(RM) -f ./*.pb.go
