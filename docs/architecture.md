# Architecture

```
client -> gRPC server -> interceptors -> service impl -> domain service -> repository -> Postgres
```

## Layers

| Layer | Package | Responsibility |
| ----- | ------- | -------------- |
| Transport | `internal/grpc` | proto <-> domain mapping, interceptors |
| Domain | `internal/domain/*` | business rules, transport-agnostic |
| Data | repositories + `internal/models` | persistence via GORM |
| Infra | `internal/infra`, `internal/events` | db, cache, event bus |

## Interceptors

They run in order: recovery, request ID, logging, auth. The auth
interceptor validates the bearer access token, loads the session, and
puts the principal in the context. Admin methods also require the admin
role.

## Errors

Domain code returns `*apperrors.HTTPError` (see `internal/errors`). The
transport layer maps them to gRPC status codes, so services never build
status errors directly.
