---
name: add-rpc
description: Add an RPC to an existing gRPC service end to end. Use when the user asks for a new endpoint, method, or API operation on an existing domain (auth, user, admin).
---

# Add an RPC to an existing service

## Steps

1. Define the RPC and its messages in `api/proto/<domain>/v1/<domain>.proto`.
   Reuse `common.v1` messages where they fit. Add a `google.api.http`
   annotation so the gateway exposes it, and `buf.validate` rules on every
   request field (email format, min_len/max_len, enums).
2. Run `just proto`. Never edit `gen/`.
3. Implement the method on the matching server in `internal/grpcsvc/`,
   mapping protobuf to domain types and calling a domain service. Handlers
   stay thin; business logic belongs in `internal/<domain>/`.
4. Return typed errors from `internal/errs` (never build a `status.Status`
   in services). New app codes go in `internal/errs/codes/`, new message
   keys in `locales/en.toml` and `locales/fr.toml`.
5. If the method is public (no bearer token), add its full method string to
   `publicMethods` in `internal/interceptor/interceptors.go`. Admin-only
   methods are covered by the `/admin.v1.AdminService/` prefix.
6. If the RPC needs new queries: add them under `internal/store/query/`,
   run `just sqlc`, and call them through the domain repository inside
   `store.Run`.
7. Add unit tests for the service logic (see `internal/auth/service_test.go`
   for the fake-collaborator pattern).
8. Add a Bruno request for it under `bruno/grpc/<domain>/` and, if it has a
   gateway route, under `bruno/gateway/<domain>/` with tests.
9. Verify: `go build ./...`, `just test`, `just lint`, `buf lint`.
