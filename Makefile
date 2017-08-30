.PHONY: all build package
REV = $(shell git rev-parse HEAD)
VERSION ?= $(shell git describe --tags --match=v* --exact-match $(REV) 2> /dev/null || echo $(REV))

all: build

build:
	CC=clang go build -ldflags "-X main.Version=$(VERSION)"

test:
	go test ./...

package: build
	-rm -rf _build
	mkdir -p _build/amd64
	mv sudolikeaboss _build/amd64/sudolikeaboss
	@cd _build/amd64; zip -r sudolikeaboss_$(VERSION)_darwin_amd64.zip .
