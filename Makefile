export GO111MODULE := on

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go mod download
.PHONY: setup

build:
	go build ./...
.PHONY: build

test:
	go test -v -failfast ./...
.PHONY: test

cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

lint:
	./bin/golangci-lint run --tests=false --enable-all --disable=lll ./...
.PHONY: lint

ci: build test lint
.PHONY: ci

.DEFAULT_GOAL := ci
