---
description: Regenerate all generated code (proto, sqlc, emails) and check drift
---

Regenerate every generated artifact and confirm the working tree only
contains expected changes:

1. `just proto`
2. `just sqlc`
3. `just emails`
4. `git status --short` on `gen/` and `templates/emails/`

Summarize what changed and why. If generation changed files that no source
change explains, investigate before committing. Never hand-edit `gen/` or
`templates/emails/`.
