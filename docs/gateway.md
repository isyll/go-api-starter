# HTTP/JSON gateway

An optional grpc-gateway reverse proxy exposes the gRPC API over
HTTP/JSON. It is off by default and never started by `just run` or the
default compose services.

## How it works

REST routes are generated entirely from the `google.api.http` annotations
on the proto RPCs, so there is no hand-maintained routing. `just proto`
regenerates the `*.pb.gw.go` stubs and a per-service OpenAPI v2 spec
under `gen/openapiv2`. The `cmd/gateway` binary dials the gRPC server and
translates HTTP/JSON to gRPC. Bearer token, `accept-language`, and
`x-request-id` headers are forwarded, so auth and i18n work through the
gateway unchanged. Client-streaming RPCs (avatar upload) are gRPC-only
and intentionally not exposed.

## Enabling it

Configure it in `configs/gateway.yaml` (env-substituted):

```yaml
enabled: ${GATEWAY_ENABLED:-false}       # must be true to run
listen_addr: ${GATEWAY_LISTEN_ADDR:-:8081}
upstream_addr: ${GATEWAY_UPSTREAM_ADDR:-} # empty -> local server port
```

```sh
GATEWAY_ENABLED=true just gateway
```

The gateway then serves REST on `:8081` and proxies to the gRPC server.
With it disabled the binary logs and exits, so it is safe to leave in a
process manager that you have not opted into yet.

## Example

```sh
curl -sX POST localhost:8081/v1/auth/login \
  -d '{"email":"a@b.com","password":"password123"}'

curl -s localhost:8081/v1/users/me \
  -H "authorization: Bearer <token>"
```
