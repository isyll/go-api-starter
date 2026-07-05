---
mode: agent
description: Add a transactional email template with the React Email design system
---

Add the requested email template (guide: docs/emails.md):

1. Create `emails/src/templates/<kebab-name>.tsx` with a copy record per
   locale (en, fr), a default-exported component built from
   `emails/src/components/`, `PreviewProps`, and `goVar("Name")` for every
   value substituted at send time.
2. Register it in `emails/src/build.tsx` with its snake_case output name.
3. Run `just emails` and commit the generated HTML.
4. Add the template constant in `internal/worker/emails/templates.go`,
   subject keys in `locales/{en,fr}.toml`, and enqueue with
   `emails.Email{TemplateID, Language, Subject, TemplateData}`.
5. Verify: `pnpm -C emails typecheck`, `just emails`,
   `go test ./internal/worker/emails/`.
