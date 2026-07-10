# API testing with Bruno

The repo ships a complete [Bruno](https://www.usebruno.com) collection in
[`bruno/`](../bruno/). Everything is plain `.bru` text, versioned with the
code, so requests evolve in the same commits as the API.

```text
bruno/
  bruno.json            collection manifest (open this folder in Bruno)
  environments/
    dev.bru             local defaults (grpc://localhost:8080, :8081)
    prod.bru            placeholders; tokens marked secret
  grpc/                 one native gRPC request per procedure
    auth/ user/ admin/ health/
  gateway/              the same procedures over the HTTP/JSON gateway
    auth/ user/ admin/ health/
    flows/              scripted end-to-end scenarios with tests
```

## Setup

1. Install the Bruno app (2.10 or newer for gRPC) and open the `bruno/`
   folder as a collection, or use the CLI: `npx @usebruno/cli`.
2. Start the stack: `just up`, `just migrate`, `just run`, and for the
   gateway requests `GATEWAY_ENABLED=true just gateway`. The worker
   (`just worker`) is only needed when a scenario exercises events or
   emails.
3. Select the `dev` environment.

## Native gRPC requests

Every RPC has a request under `grpc/`, including the client-streaming
`UploadAvatar`. Requests use **server reflection** (enabled by default in
development through `GRPC_REFLECTION`), so no proto import is needed:
open a request, hit send, and Bruno resolves the schema from the server.

Authenticated procedures send `authorization: Bearer {{access_token}}`
metadata. Run `gateway/flows` or the gateway Login request first; their
post-response scripts store `access_token` and `refresh_token` as
runtime variables.

> [!NOTE]
> Bruno does not run scripts or tests on gRPC requests yet (planned for
> its next releases), which is why assertions and automation live on the
> gateway side. The gateway transcodes to the same RPCs, so both paths
> exercise identical handlers, interceptors, and validation.

## Gateway requests and tests

Each request under `gateway/` mirrors one RPC over HTTP and carries
tests: an expected-status check and a no-server-error guard. Requests
that need state you may not have (a real verification token, an admin
account) document the deterministic failure they assert instead.

## Automation flows

`gateway/flows` is a numbered end-to-end scenario meant to run as a
suite: register a unique account, prove the token works, rotate the
refresh token, prove reuse detection revokes the family, recover by
logging in, reject invalid payloads, change the password, log in with
the new one, log out, and prove the revoked token is dead. Every step
asserts exact status codes and response shapes, and chains state through
runtime variables.

Run it from the app (Collection Runner on the `flows` folder) or the
CLI:

```sh
just api-test
```

The flows create a fresh account per run, so they are idempotent and
safe to run repeatedly against a development stack.

## Environments and secrets

`dev.bru` contains local defaults only. `prod.bru` declares
`access_token` and `refresh_token` under `vars:secret`, so Bruno keeps
their values out of the committed file; fill endpoint URLs per
deployment. Scripts store tokens with `bru.setVar` (runtime only, never
written to disk).

## Adding requests

New RPC (see the add-rpc skill): add one `.bru` under the matching
`grpc/<domain>/` and `gateway/<domain>/` folder, reusing an existing
file as the template, and extend a flow when the RPC participates in a
scenario worth automating.

---

**See also:** [gRPC API](grpc.md) · [HTTP/JSON gateway](gateway.md)
