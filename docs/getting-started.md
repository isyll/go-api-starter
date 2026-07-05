# Getting started

## Prerequisites

- Go 1.26+
- Docker (for Postgres and Redis)
- [just](https://github.com/casey/just) and [buf](https://buf.build)

## Setup

```sh
cp .env.example .env
just up          # start postgres + redis + minio
just migrate     # apply migrations
just run         # start the gRPC server on :8080
```

Metrics and health are on `:9090` (`/metrics`, `/healthz`, `/readyz`).

## Calling the API

The server enables gRPC reflection, so grpcurl can explore it:

```sh
grpcurl -plaintext localhost:8080 list

grpcurl -plaintext -d '{"email":"a@b.com","password":"password123","first_name":"A","last_name":"B"}' \
  localhost:8080 auth.v1.AuthService/Register
```

Authenticated calls pass the access token as metadata:

```sh
grpcurl -plaintext -H "authorization: Bearer <token>" \
  localhost:8080 user.v1.UserService/GetMe
```

To reach the API over HTTP/JSON instead, enable the optional
[gateway](gateway.md).
