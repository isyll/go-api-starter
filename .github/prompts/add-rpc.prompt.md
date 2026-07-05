---
mode: agent
description: Add an RPC to an existing gRPC service end to end
---

Add the requested RPC following the template conventions:

1. Define the RPC in `api/proto/<domain>/v1/<domain>.proto` with a
   `google.api.http` annotation and `buf.validate` rules on request
   fields, then run `just proto`.
2. Implement the method in `internal/grpcsvc/`, delegating business logic
   to the domain service in `internal/<domain>/`.
3. Return typed errors from `internal/errs`; add app codes in
   `internal/errs/codes/` and message keys in `locales/{en,fr}.toml`.
4. Public methods must be listed in `publicMethods` in
   `internal/interceptor/interceptors.go`.
5. New queries go in `internal/store/query/` followed by `just sqlc`.
6. Add service tests (see `internal/auth/service_test.go`) and Bruno
   requests under `bruno/grpc/` and `bruno/gateway/`.
7. Verify: `go build ./...`, `just test`, `just lint`, `buf lint`.
