---
description: Run the full local verification suite and report failures
---

Run the same checks CI runs, in this order, and fix anything that fails:

1. `gofmt -l cmd internal pkg` (must print nothing)
2. `go vet ./...`
3. `go build ./...`
4. `just test`
5. `just lint`
6. `buf lint`
7. `buf generate && git diff --exit-code gen/`
8. `sqlc generate && git diff --exit-code gen/db`
9. `pnpm -C emails typecheck && pnpm -C emails build && git diff --exit-code templates/emails`

Report each step as pass or fail. If a step fails, show the error, fix the
root cause (never mask it), and rerun from that step.
