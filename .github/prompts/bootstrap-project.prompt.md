---
mode: agent
description: Turn this template into a new project (module path, branding, identifiers)
---

Bootstrap a new project from this template. Ask for the Go module path,
app name, and web URL if not provided, then:

1. Replace `github.com/isyll/go-grpc-starter` with the new module path in
   `go.mod`, all imports, `buf.gen.yaml`, `justfile`, and
   `infra/docker/Dockerfile`.
2. Set the app name in `.env.example`, `configs/api.yaml`, and
   `errorDomain` in `internal/interceptor/errors.go`.
3. Generate a fresh unique sqids alphabet for `ID_OBFUSCATION_ALPHABET`.
4. Rebrand `emails/src/theme/tokens.ts`, then run `just emails`.
5. Regenerate and verify: `just proto`, `just sqlc`, `go build ./...`,
   `just test`, `just lint`.

Never edit `gen/` or `templates/emails/` by hand.
