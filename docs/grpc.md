# gRPC API

The API is defined in `api/proto/<domain>/v1`, one protobuf package per
domain (`common.v1`, `auth.v1`, `user.v1`, `admin.v1`, `health.v1`).
Shared messages live in `common.v1`. Generated Go code lives in `gen`
and must not be edited by hand.

## Regenerate

```sh
just proto   # buf lint && buf generate
```

## Adding an RPC

1. Add the RPC and its messages to the `.proto` for that domain, reusing
   `common.v1` messages where they fit.
2. Run `just proto`.
3. Implement the method on the matching server in `internal/grpcsvc`
   (map proto <-> domain), and call a domain service for the work.
4. Return typed errors from `internal/errs`; never build a `status.Status`
   in the handler or service.

## Adding a new service

1. Create `api/proto/<domain>/v1/<domain>.proto` with `package <domain>.v1`
   and a service; import `common/v1/common.proto` for shared messages.
2. Run `just proto`.
3. Add a server type in `internal/grpcsvc` that embeds the generated
   `Unimplemented<Name>ServiceServer`, and register it in `server.go`.
4. If the service (or some of its methods) is public or admin-only,
   update `publicMethods` / `adminServicePrefix` in
   `internal/interceptor/interceptors.go` with the new
   `/<domain>.v1.<Service>/<Method>` route strings.

## Services

| Service | Package | Purpose |
| ------- | ------- | ------- |
| `HealthService` | `health.v1` | liveness and readiness |
| `AuthService` | `auth.v1` | register, login, tokens, devices, password reset |
| `UserService` | `user.v1` | own profile, settings, avatar upload, push tokens |
| `AdminService` | `admin.v1` | admin-only user management (admin role) |

> [!NOTE]
> Public methods (register, login, refresh, verify/reset) are listed in
> `internal/interceptor/interceptors.go`. Everything else requires a token.

For optional REST/JSON access to these RPCs, see
[the gateway](gateway.md).
