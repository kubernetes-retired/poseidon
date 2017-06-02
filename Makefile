# Disable make's implicit rules, which are not useful for golang, and slow down the build
# considerably.
.SUFFIXES:


GO_PATH=$(GOPATH)
SRCFILES=poseidon.go
TEST_SRCFILES=$(wildcard *_test.go)

# Ensure that the dist directory is always created
MAKE_SURE_DIST_EXIST := $(shell mkdir -p dist)

.PHONY: all plugin
default: clean all
all: plugin
plugin: dist/poseidon

.PHONY: test
test: dist/poseidon-test

.PHONY: clean
clean:
	rm -rf dist

release: clean

# Build the poseidon
dist/poseidon: $(SRCFILES)
	@GOPATH=$(GO_PATH) CGO_ENABLED=0 go build -v -i -o dist/poseidon \
	-ldflags "-X main.VERSION=1.0 -s -w" poseidon.go

# Build the poseidon tests
dist/poseidon-test: $(TEST_SRCFILES)
	@GOPATH=$(GO_PATH) CGO_ENABLED=0 ETCD_IP=127.0.0.1 go test
