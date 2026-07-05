import { render } from "@react-email/render";
import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";
import type { ComponentType } from "react";
import PasswordReset, { passwordResetCopy } from "./templates/password-reset";
import VerifyEmail, { verifyEmailCopy } from "./templates/verify-email";
import { locales, type Locale } from "./vars";

// Each entry renders to templates/emails/<name>/<locale>.html, where <name>
// must match the TemplateID the Go worker uses.
type Entry = {
  name: string;
  Component: ComponentType<{ copy: never; locale: Locale }>;
  copy: Record<Locale, unknown>;
};

const registry: Entry[] = [
  { name: "verify_email", Component: VerifyEmail as Entry["Component"], copy: verifyEmailCopy },
  { name: "password_reset", Component: PasswordReset as Entry["Component"], copy: passwordResetCopy },
];

const outDir = path.resolve(import.meta.dirname, "../../templates/emails");

for (const { name, Component, copy } of registry) {
  const dir = path.join(outDir, name);
  await mkdir(dir, { recursive: true });
  for (const locale of locales) {
    const html = await render(
      <Component copy={copy[locale] as never} locale={locale} />,
      { pretty: true },
    );
    const file = path.join(dir, `${locale}.html`);
    await writeFile(file, html + "\n");
    console.log(`rendered ${name}/${locale}.html`);
  }
}
