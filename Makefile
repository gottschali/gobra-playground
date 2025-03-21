.PHONY: build test vet lint fmt clean
OUT := gobra-playground
PKG := github.com/gottschali/gobra-playground
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: build

build:
	@go build

test:
	@go test -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	golangci-lint run

fmt:
	@gofmt -l -w -s ${GO_FILES}

clean:
	-@rm ${OUT}
