# Project guide

A Go backend service exposing a gRPC API, backed by PostgreSQL, Redis,
and background workers.

## Stack

- Go with gRPC (protobuf, generated with buf)
- PostgreSQL via GORM, with row-level security
- Redis for caching and opaque access tokens
- Asynq workers for email, push, and event dispatch
- golang-migrate for SQL migrations

## Layout

```text
cmd/            entrypoints: server, migrate, worker/*
proto/          protobuf definitions (source of the API)
internal/
  grpc/         gRPC server, interceptors, service implementations
  domain/       business logic per area (auth, users, ...)
  events/       event bus + transactional outbox
  worker/       Asynq worker processors
  infra/        db (RLS), cache
  models/       GORM models
  gen/          generated protobuf code (do not edit)
migrations/     SQL migrations (golang-migrate)
configs/        YAML config, read with env substitution
infra/          docker, postgres init, prometheus
```

## Common commands

```sh
just run            # run the gRPC server
just test           # run tests
just proto          # regenerate protobuf code
just migrate        # apply migrations
just up             # start postgres + redis (docker)
just lint           # golangci-lint
```

## Conventions

- Services return `*apperrors.HTTPError`; the gRPC error interceptor
  maps them to status codes. Don't build status errors in services.
- Never edit files under `gen`. Change the `.proto` and run
  `just proto`.
- New API: add the RPC to a `.proto`, run `just proto`, implement the
  method in `internal/grpc`, and call a domain service.
- Keep comments short. Explain why, not what.
