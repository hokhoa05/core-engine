.PHONY: build run test clean

APP_NAME=core-engine

build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) cmd/engine/main.go

run: build
	@echo "Running $(APP_NAME)..."
	@./bin/$(APP_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf bin/