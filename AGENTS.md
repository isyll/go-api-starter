# Project guide

> A gRPC-native Go backend template, backed by PostgreSQL, Redis, and
> background workers. Every layer is designed for gRPC from day one.

Deep-dive documentation lives in [`docs/`](docs/README.md); this file is
the quick reference for working in the repo.

## Stack

- Go with gRPC (protobuf, generated with buf)
- PostgreSQL via pgx + sqlc, with row-level security
- Redis for caching and opaque access tokens
- Asynq workers for email, push, and event dispatch
- golang-migrate for SQL migrations
- Optional grpc-gateway HTTP/JSON transcoding (off by default)

## Layout

```text
cmd/            entrypoints: server, gateway (opt-in), migrate, worker
api/proto/      protobuf definitions, one package per domain
                (common.v1, auth.v1, user.v1, admin.v1, health.v1)
internal/
  grpcsvc/      thin gRPC handlers: protobuf <-> domain mapping
  interceptor/  unary interceptors (recovery, logging, locale,
                error-map, request-id, auth)
  <domain>/     business logic per area; entities live here too
  store/        pgx + sqlc engine; RLS-scoped transactions
  event/        event bus + transactional outbox
  worker/       Asynq worker processors (emails, notifications)
  platform/     db (pgx pool + RLS), cache (redis), storage (minio)
  errs/         typed gRPC-native errors
  reqctx/       per-request context (language, request id, auth subject)
gen/            generated code, do not edit: <domain>/v1, db, openapiv2
migrations/     SQL migrations (golang-migrate)
configs/        YAML config, read with env substitution
locales/        i18n message catalogs (go-i18n)
infra/          docker, postgres init, prometheus
```

## Commands

| Command | Action |
| --- | --- |
| `just run` | run the gRPC server |
| `just gateway` | run the optional HTTP/JSON gateway (opt-in) |
| `just worker` | run the background workers |
| `just test` | run tests |
| `just proto` | regenerate protobuf code (buf) |
| `just sqlc` | regenerate sqlc code |
| `just migrate` | apply migrations |
| `just up` | start postgres + redis + minio (docker) |
| `just lint` | golangci-lint |

## Conventions

- Services return typed errors from `internal/errs` (a gRPC `codes.Code`
  plus a machine app code and an i18n message key). The error
  interceptor is the only place that builds a `status.Status` with
  details. Don't build status errors in services, and keep `net/http`
  out of domain and error code.
- Never edit files under `gen`. Change the `.proto` and run `just proto`,
  or the queries under `internal/store/query` and run `just sqlc`.
- New API: add the RPC to a `.proto` under `api/proto/<domain>/v1`, run
  `just proto`, implement the method in `internal/grpcsvc`, and call a
  domain service. See [docs/grpc.md](docs/grpc.md).
- Keep comments short. Explain why, not what.

> [!NOTE]
> `CLAUDE.md` is a symlink to this file, so both stay in sync.
