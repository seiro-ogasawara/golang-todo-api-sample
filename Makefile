.PHONY: test
.SILENT: migrate test

build:
	go build -o api-server main.go

LINT_FILES = ./...
lint:
	golangci-lint run --timeout=3m0s $(LINT_FILES)

DB_PASSWORD ?= postgres
DB_USER ?= postgres
DB_NAME ?= test
DB_PORT ?= 15432
DB_HOST ?= localhost
CI ?= false

migrate:
ifneq ($(CI),true)
	migrate -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable \
		-path migration up
endif

test: migrate
	DB_PASSWORD=$(DB_PASSWORD) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) \
	go test ./...
