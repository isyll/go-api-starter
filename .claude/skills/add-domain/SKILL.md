---
name: add-domain
description: Create a new domain service (proto package, migration, queries, repository, service, gRPC server, wiring). Use when the user asks for a whole new feature area or bounded context.
---

# Add a new domain service

Example domain name used below: `billing`.

## Steps

1. Proto: create `api/proto/billing/v1/billing.proto` with
   `package billing.v1`, a `service BillingService`, `google.api.http`
   annotations, and `buf.validate` rules. Run `just proto`; generated code
   lands in `gen/billing/v1`.
2. Migration: `just migrate-create create_billing_tables`, write up and down
   SQL under `migrations/`. Use `timestamptz`, `NOT NULL` where possible,
   and indexes for every foreign key and query filter. Add RLS policies if
   rows are user-scoped (see existing tables for the GUC pattern).
3. Queries: add `internal/store/query/billing.sql` with sqlc annotations,
   then `just sqlc`.
4. Domain package `internal/billing/`:
   - `entity.go` (domain types), `errors.go` (typed errors from
     `internal/errs` with new codes in `internal/errs/codes/billing.go`),
   - `repository.go` (wraps `store.Run`, maps rows to entities, returns
     errors, never panics),
   - `service.go` (business rules; wrap multi-write flows in
     `store.WithTx` through a TxRunner collaborator).
5. gRPC server: `internal/grpcsvc/billing_service.go` embedding
   `billingv1.UnimplementedBillingServiceServer`; register it in
   `internal/grpcsvc/server.go` and wire dependencies in
   `internal/app/wire.go`.
6. Auth: public methods go in `publicMethods`
   (`internal/interceptor/interceptors.go`); everything else requires a
   bearer token automatically.
7. Locales: add message keys to `locales/en.toml` and `locales/fr.toml`.
8. Events: if the domain emits events, follow the add-event skill.
9. Tests: service tests with fake collaborators, repository conversion
   tests where useful.
10. Bruno: add request folders under `bruno/grpc/billing/` and
    `bruno/gateway/billing/`.
11. Verify: `just proto && just sqlc`, `go build ./...`, `just test`,
    `just lint`, `git diff --exit-code gen/`.
