#!/bin/sh

snap install go --classic
apt install -y llvm clang libbpf-dev protobuf-compiler npm
go install google.golang.org/protobuf/cmd/protoc-gen-go
make
