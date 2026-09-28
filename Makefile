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

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) internal/web/dist
