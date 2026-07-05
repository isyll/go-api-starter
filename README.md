# go-grpc-starter

A reusable, gRPC-native Go backend template: a gRPC API, PostgreSQL with
row-level security, Redis cache, a transactional outbox event system,
background workers, and an optional HTTP/JSON gateway.

## Stack

- Go + gRPC (protobuf, buf)
- PostgreSQL via pgx + sqlc, with row-level security
- Redis (cache + opaque access tokens)
- Asynq workers (email, push, event dispatcher)
- golang-migrate migrations
- Optional grpc-gateway HTTP/JSON transcoding (off by default)

## Getting started

```sh
cp .env.example .env
just up        # start postgres + redis + minio
just migrate   # apply migrations
just run       # run the gRPC server
```

See [docs/](docs/) for architecture, the API, auth, the database, events,
and how to enable the gateway.
