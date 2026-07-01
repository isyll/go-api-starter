set dotenv-load
set dotenv-filename := ".env"

app_name := "app_owner-api"
build_dir := "build"

# Run

# Run API server
run:
    @go run ./cmd/api

# Run API with hot reload (air)
dev:
    @rm -rf tmp/ && air

# Code quality

# Format all Go code (gofumpt + goimports + golines) and SQL migrations
format:
    @gofumpt -w . && goimports -w . && gofmt -s -w .
    @golines -w -m 72 --no-reformat-tags \
        --base-formatter=gofumpt .
    @go mod tidy
    @sqlfluff fix --dialect postgres migrations/

# Lint and auto-fix (golangci-lint + staticcheck full ruleset)
lint:
    @golangci-lint run --fix
    @staticcheck -checks=all ./...

# Run staticcheck with the strictest profile (-checks=all)
staticcheck:
    @staticcheck -checks=all ./...

# Generate Swagger docs
swagger:
    @swag init -g cmd/api/main.go -o swagger \
        --outputTypes go,yaml,json

# Run the standalone Swagger UI server
swagger-serve:
    @go run ./cmd/swagger

# Tests

# Run all tests
test:
    @go test ./... -race -shuffle=on -timeout 30s

# Run tests with coverage report
test-cover:
    @go test ./... -race -shuffle=on -coverprofile=coverage.out \
        -covermode=atomic -timeout 30s
    @go tool cover -func=coverage.out

# Run unit tests only
test-unit:
    @go test ./internal/helpers/... ./internal/models/... \
        ./pkg/utils/... ./pkg/api/... \
        ./pkg/validators/... -v -race -timeout 30s

# Run integration tests only
test-integration:
    @go test ./internal/handlers/... \
        -v -race -timeout 60s

# Migrations

# Run all migrations up
migrate-up:
    @go run ./cmd/migrate -up

# Rollback migrations (defaults to 1 step)
migrate-down steps="1":
    @go run ./cmd/migrate -down -steps {{ steps }}

# Create migration
migrate-create name:
    @migrate create -ext sql -dir migrations -seq {{ name }}

# Force version
migrate-force version:
    @go run ./cmd/migrate -force {{ version }}

# Build

# Build API binary
build:
    @mkdir -p {{ build_dir }}
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/{{ app_name }} ./cmd/api

# Build standalone Swagger UI binary
swagger-build:
    @mkdir -p {{ build_dir }}
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-swagger ./cmd/swagger

# Build API + migrate + all workers
build-all:
    @mkdir -p {{ build_dir }}
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/{{ app_name }} ./cmd/api
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-swagger ./cmd/swagger
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-migrate ./cmd/migrate
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-push-worker \
        ./cmd/worker/push_notifications
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-email-worker \
        ./cmd/worker/email_sender
    @go build -trimpath -ldflags="-s -w" \
        -o {{ build_dir }}/app_owner-trip-worker \
        ./cmd/worker/trip_lifecycle

# Workers

# Run push notifications worker
worker-push:
    @go run ./cmd/worker/push_notifications

# Run email worker
worker-email:
    @go run ./cmd/worker/email_sender

# Run trip lifecycle worker
worker-trips:
    @go run ./cmd/worker/trip_lifecycle

# Scripts

# Bump API version: patch | minor | major | X.Y.Z
bump v:
    @chmod +x scripts/bump_version.sh && ./scripts/bump_version.sh {{ v }}

# Manage pg_cron session cleanup job
pg-cron +args:
    @chmod +x scripts/manage-pg-cron.sh && ./scripts/manage-pg-cron.sh {{ args }}

# Renumber migration files from a given start number
renumber-migrations start +args="":
    @chmod +x scripts/renumber_migrations.sh \
        && ./scripts/renumber_migrations.sh {{ start }} {{ args }}

# Run the trip schedule generator once
ride-generator:
    @chmod +x scripts/run-ride-generator.sh && ./scripts/run-ride-generator.sh

# Cleanup

# Clean build artifacts
clean:
    @rm -rf {{ build_dir }} tmp/ swagger/ logs/ \
        secrets/ coverage.out
    @go clean
