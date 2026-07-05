# Architecture

```text
client -> gRPC server -> interceptors -> grpcsvc handler -> domain service -> store (pgx+sqlc) -> Postgres
```

## Layers

| Layer | Package | Responsibility |
| ----- | ------- | -------------- |
| Transport | `internal/grpcsvc`, `internal/interceptor` | proto <-> domain mapping, interceptors |
| Domain | `internal/<domain>` | business rules, transport-agnostic; entities live here |
| Data | `internal/store` (pgx + sqlc) + per-domain repositories | persistence in RLS-scoped transactions |
| Platform | `internal/platform`, `internal/event` | db pool, cache, object storage, event bus |

## Interceptors

They run outermost to innermost: recovery, logging, locale, error-map,
request-id, auth. The auth interceptor validates the bearer access
token, loads the session, and puts the authenticated subject in the
context (`internal/reqctx`). Admin methods also require the admin role.
Public methods are listed in `internal/interceptor/interceptors.go`.

## Errors

Domain code returns typed errors from `internal/errs` (a grpc
`codes.Code`, a machine app code, and an i18n message key). The error
interceptor is the only place that turns them into a localized
`status.Status` with details, so services never build status errors
directly and no HTTP status codes appear anywhere.
