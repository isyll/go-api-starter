# gRPC API

The API is defined in `proto/api/v1`. Generated Go code lives in
`internal/gen` and must not be edited by hand.

## Regenerate

```sh
just proto   # buf lint && buf generate
```

## Adding an endpoint

1. Add the RPC and messages to a `.proto` file.
2. Run `just proto`.
3. Implement the method in `internal/grpc` (map proto <-> domain).
4. Call a domain service for the actual work.

## Services

| Service | Purpose |
| ------- | ------- |
| `HealthService` | liveness and readiness |
| `AuthService` | register, login, tokens, devices, password reset |
| `UserService` | own profile, settings, push tokens |
| `AdminService` | admin-only user management (admin role) |

> [!NOTE]
> Public methods (register, login, refresh, verify/reset) are listed in
> `internal/grpc/interceptors.go`. Everything else requires a token.
