# Static site development. Go is the only required build tool.
OUT       ?= dist
BASE_PATH ?=
PORT      ?= 8080
AIR       := $(shell command -v air 2>/dev/null)

.DEFAULT_GOAL := build
.PHONY: build run preview dev fmt vet tidy test check tools help

## build: generate the deployable site (OUT=dist BASE_PATH=)
build:
	@go run ./cmd/build -out "$(OUT)" -base-path "$(BASE_PATH)"

## run: build and preview locally (PORT=8080)
run: build
	@$(MAKE) --no-print-directory preview

## preview: serve the existing output without rebuilding
preview:
	@go run ./cmd/server -dir "$(OUT)" -base-path "$(BASE_PATH)" -addr ":$(PORT)"

## dev: rebuild/restart on edits with optional Air, or build and preview once
dev:
ifeq ($(AIR),)
	@echo "Air not installed; building and previewing once. Use make tools for watch mode."
	@$(MAKE) --no-print-directory run
else
	@OUT="$(OUT)" BASE_PATH="$(BASE_PATH)" PORT="$(PORT)" air -c .air.toml
endif

## fmt: format Go sources
fmt:
	@gofmt -w -s assets.go cmd internal

## vet: run static checks
vet:
	@go vet ./...

## tidy: sync module metadata
tidy:
	@go mod tidy

## test: run the test suite
test:
	@go test ./...

## check: format, vet, and test
check: fmt vet test

## tools: install optional Air file watcher
tools:
	@go install github.com/air-verse/air@latest

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
