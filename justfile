set dotenv-load
set dotenv-filename := ".env"

build_dir := "bin"

# List available recipes.
default:
    @just --list

# Run the gRPC server.
run:
    @go run ./cmd/server

# Run with hot reload (air).
dev:
    @air

# Generate protobuf code (buf).
proto:
    @buf lint && buf generate

# Generate sqlc code from migrations + queries.
sqlc:
    @sqlc generate

# Format Go code.
fmt:
    @gofmt -w cmd internal pkg
    @go mod tidy

# Lint (golangci-lint).
lint:
    @golangci-lint run

# Run tests.
test:
    @go test ./... -race -shuffle=on -timeout 60s

# Tests with coverage.
test-cover:
    @go test ./... -race -coverprofile=coverage.out -covermode=atomic
    @go tool cover -func=coverage.out

# Build all binaries.
build:
    @mkdir -p {{ build_dir }}
    @go build -trimpath -ldflags="-s -w" -o {{ build_dir }}/ ./cmd/...

# Apply all migrations.
migrate:
    @go run ./cmd/migrate -up

# Roll back N migrations (default 1).
migrate-down steps="1":
    @go run ./cmd/migrate -down -steps {{ steps }}

# Create a new migration pair.
migrate-create name:
    @migrate create -ext sql -dir migrations -seq {{ name }}

# Migration status.
migrate-status:
    @go run ./cmd/migrate -status

# Run all background workers (email, push, event dispatcher).
worker:
    @go run ./cmd/worker

# Start postgres + redis + minio with docker.
up:
    @docker compose up -d postgres redis minio

# Start the full stack with docker.
up-all:
    @docker compose up -d --build

# Stop the docker stack.
down:
    @docker compose down

# Clean build artifacts.
clean:
    @rm -rf {{ build_dir }} tmp/ coverage.out
    @go clean
