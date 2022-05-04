.PHONY: test

build:
	go build -o api-server main.go

LINT_FILES = ./...
lint:
	golangci-lint run --timeout=3m0s $(LINT_FILES)

test:
	go test ./...
