# go-api-starter

A reusable Go backend template: gRPC API, PostgreSQL (GORM + row-level security),
Redis cache, a transactional outbox event system, and background workers.

## Stack

- Go + gRPC (protobuf, buf)
- PostgreSQL with GORM and row-level security
- Redis (cache + opaque access tokens)
- Asynq workers (email, push, event dispatcher)
- golang-migrate migrations

## Getting started

```sh
cp .env.example .env
just up        # start postgres + redis
just migrate   # apply migrations
just run       # run the gRPC server
```

See [docs/](docs/) for more.
