GO ?= go
BINARY ?= bin/big-brother

.PHONY: dev build web test vet lint sqlc clean run

## build: compile the web app and the single binary (real UI embedded)
build: web
	$(GO) build -o $(BINARY) ./cmd/server

## web: build the Vite app into the embedded dist directory
web:
	cd web && npm install && npm run build
	@touch internal/web/dist/.gitkeep

## dev: run the Go server (expects `make web` for the UI, or run vite separately)
dev:
	$(GO) run ./cmd/server

## run: build and run the binary
run: build
	./$(BINARY)

## test: run all Go tests
test:
	$(GO) test ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## lint: vet plus formatting check
lint: vet
	@test -z "$$(gofmt -l cmd internal)" || (echo "gofmt needed:"; gofmt -l cmd internal; exit 1)

## sqlc: regenerate typed queries
sqlc:
	sqlc generate

## openapi: regenerate the TypeScript client from the OpenAPI document
openapi:
	cd web && npm run gen:api

## openapi-check: fail if the generated client is stale
openapi-check:
	cd web && npm run gen:api
	@git diff --exit-code -- web/src/lib/api/generated || \
		(echo "generated API client is stale: run 'make openapi'"; exit 1)

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) internal/web/dist
