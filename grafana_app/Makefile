VENDOR_DIR := vendor
GOBIN_DIR := $(shell go env GOPATH)/bin
JSONNET_VENDOR_DIR := $(VENDOR_DIR)/jsonnet

### Dependencies
deps-dashboard:
	go install github.com/google/go-jsonnet/cmd/jsonnet@latest

$(JSONNET_VENDOR_DIR): jsonnetfile.json
	mkdir -p $(JSONNET_VENDOR_DIR)
	go install -a github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest
	$(GOBIN_DIR)/jb --jsonnetpkg-home="$(JSONNET_VENDOR_DIR)" install

node_modules: package.json
	npm install --global yarn
	yarn install

deps-frontend: node_modules

### Build

dist/%.json: src/%.jsonnet deps-dashboard $(JSONNET_VENDOR_DIR)
	echo $(shell pwd)
	mkdir -p $(dir $@)
	$(GOBIN_DIR)/jsonnet -J "$(JSONNET_VENDOR_DIR)" -o $@ $<

build-dashboards: deps-dashboard $(shell find src -name '*.jsonnet' | sed -E 's@src/(.+)\.jsonnet@dist/\1.json@g')

build-frontend: deps-frontend
	yarn run build

GO_LD_FLAGS := -w -s -extldflags "-static"
build-backend:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/datasource/gpx_grafana_hermes_datasource_linux_amd64 -ldflags '$(GO_LD_FLAGS)' ./pkg

build: build-frontend build-backend build-dashboards

### Clean
clean:
	rm -rf node_modules "$(VENDOR_DIR)" dist build
