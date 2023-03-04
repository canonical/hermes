.PHONY: all build clean

GO     := go
RM     := rm -f
PROTOC := protoc

all: clean build

build:
	PATH=$(PATH):$(shell $(GO) env GOPATH)/bin $(PROTOC) -I ./ --go_out=./ --go_opt=module=hermes ./*.proto

clean:
	$(RM) ./*.pb.go
