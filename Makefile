.PHONY: build test lint ci

build:
	go build -v ./...

test:
	go test -v -cover ./...

lint:
	golangci-lint run

ci: build test lint
