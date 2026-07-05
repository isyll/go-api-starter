---
name: add-email-template
description: Add a transactional email template with the React Email design system. Use when the user wants a new email (welcome, invoice, alert) or a new locale for existing emails.
---

# Add an email template

Templates are React components in `emails/src/templates/`, compiled to
static HTML in `templates/emails/<name>/<locale>.html`, and sent by the Go
worker through Resend. Full guide: `docs/emails.md`.

## Steps

1. Create `emails/src/templates/<kebab-name>.tsx`:
   - export a copy record per locale (`en`, `fr`),
   - default-export the component built from `emails/src/components/`
     (EmailLayout, PrimaryButton, Title, Paragraph, Muted, FallbackLink),
   - set `PreviewProps` so `just emails-dev` can preview it,
   - use `goVar("Name")` for every value substituted at send time; it
     renders as a `{{.Name}}` Go template action.
2. Register it in `emails/src/build.tsx` with its snake_case output name.
3. Run `just emails` and commit the generated HTML (CI fails on drift).
4. Go side:
   - add the template name constant in
     `internal/worker/emails/templates.go` (must match the output dir),
   - enqueue with `emails.Email{TemplateID, Language, Subject,
     TemplateData}`; subjects come from `locales/*.toml` keys, and
     `TemplateData` keys must match the `goVar` names plus `Year`.
5. Never edit `templates/emails/` by hand; change tokens in
   `emails/src/theme/tokens.ts` for visual changes and rebuild.
6. Verify: `pnpm -C emails typecheck`, `just emails`,
   `go test ./internal/worker/emails/`.
