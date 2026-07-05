---
name: bootstrap-project
description: Turn this template into a new project. Use when the user wants to start, rename, or rebrand a project based on go-grpc-starter (new module path, app name, branding, identifiers).
---

# Bootstrap a new project from this template

Collect from the user first: the Go module path (for example
`github.com/acme/orders-api`), the app name, and the web front-end URL.

## Steps

1. Rename the Go module everywhere:
   - `go.mod` module line.
   - Every import: `grep -rl "github.com/isyll/go-grpc-starter" --include="*.go" --include="*.yaml" --include="*.md" . | xargs sed -i "s|github.com/isyll/go-grpc-starter|<new module>|g"`
   - `buf.gen.yaml` (go_package_prefix), `justfile` (module variable),
     `infra/docker/Dockerfile` (ldflags paths).
2. Set the app identity:
   - `APP_NAME` in `.env.example` and `configs/api.yaml` default.
   - `errorDomain` in `internal/interceptor/errors.go`.
3. Generate a fresh sqids alphabet (shuffled lowercase+digits, unique chars)
   and set `ID_OBFUSCATION_ALPHABET` in `.env.example`. IDs must differ
   between deployments.
4. Rebrand emails: edit `emails/src/theme/tokens.ts` (brand name, logo URL,
   palette), then `just emails`.
5. Update `README.md` title and description; keep `AGENTS.md` structure.
6. Regenerate and verify: `just proto && just sqlc && just emails`,
   `go build ./...`, `just test`, `just lint`.
7. Reset history if this is a fresh repo: `rm -rf .git && git init` then an
   initial commit. Keep history when forking instead.

## Rules

- Never edit `gen/` or `templates/emails` by hand.
- All secrets stay in `.env`; `.env.example` gets placeholders only.
- Commit messages: short, plain, imperative subjects.
