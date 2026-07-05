---
mode: agent
description: Create a new domain service (proto, migration, repository, service, wiring)
---

Create the requested domain following the template layout:

1. `api/proto/<domain>/v1/<domain>.proto` with validation and gateway
   annotations, then `just proto`.
2. Migration via `just migrate-create`, with timestamptz columns,
   constraints, indexes on foreign keys, and RLS policies for user-scoped
   tables.
3. Queries in `internal/store/query/<domain>.sql`, then `just sqlc`.
4. Domain package `internal/<domain>/` with entity, errors (typed, codes
   in `internal/errs/codes/`), repository (returns errors, uses
   `store.Run`), and service (multi-write flows in `store.WithTx`).
5. gRPC server in `internal/grpcsvc/`, registered in `server.go`, wired
   in `internal/app/wire.go`.
6. Locale keys in `locales/{en,fr}.toml`, service tests with fakes,
   Bruno requests.
7. Verify: `go build ./...`, `just test`, `just lint`, `buf lint`.
