# Getting started

> Run the stack locally and make your first call.

## Prerequisites

| Tool | Purpose |
| --- | --- |
| Go 1.26+ | build and run the services |
| Docker | Postgres, Redis, and MinIO |
| [just](https://github.com/casey/just) | task runner |
| [buf](https://buf.build) | protobuf generation |
| Node 24+ and [pnpm](https://pnpm.io) | only for editing email templates |

## Setup

```sh
cp .env.example .env
just up          # start postgres + redis + minio
just migrate     # apply migrations
just run         # start the gRPC server on :8080
```

The running system is three processes over shared infrastructure; the
worker and gateway are separate binaries:

```mermaid
flowchart LR
    SV["server · :8080"]
    WK["worker"]
    GW["gateway · :8081<br/>(opt-in)"]

    SV --> PG[("Postgres")]
    SV --> RD[("Redis")]
    WK --> PG
    WK --> RD
    GW -->|gRPC| SV
```

Run the worker with `just worker`, and the optional gateway with
`GATEWAY_ENABLED=true just gateway`.

## Call the API

The server enables gRPC reflection by default in development
(`GRPC_REFLECTION`, disable it on internet-facing deployments), so
`grpcurl` can explore it:

```sh
grpcurl -plaintext localhost:8080 list

grpcurl -plaintext \
  -d '{"email":"a@b.com","password":"password123","first_name":"A","last_name":"B"}' \
  localhost:8080 auth.v1.AuthService/Register
```

Authenticated calls pass the access token as metadata:

```sh
grpcurl -plaintext -H "authorization: Bearer <token>" \
  localhost:8080 user.v1.UserService/GetMe
```

> [!TIP]
> Prefer REST? Enable the [HTTP/JSON gateway](gateway.md) and call the
> same RPCs over `curl`.

## Observability

The API serves operational endpoints on `:9090` (`METRICS_PORT`) and
the worker on `:9091` (`WORKER_METRICS_PORT`):

| Endpoint | Purpose |
| --- | --- |
| `/metrics` | Prometheus metrics (RPCs, outbox, cache, DB and Redis pools) |
| `/healthz` | liveness |
| `/readyz` | readiness (checks Postgres + Redis) |
| `/debug/pprof/` | profiling, only when `PPROF_ENABLED=true` |

---

**See also:** [Architecture](architecture.md) · [gRPC API](grpc.md)
