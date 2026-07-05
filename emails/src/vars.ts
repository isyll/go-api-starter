// goVar emits a Go template action. The worker executes the generated HTML
// with html/template, substituting these placeholders at send time.
export function goVar(name: string): string {
  return `{{.${name}}}`;
}

export type Locale = "en" | "fr";

export const locales: Locale[] = ["en", "fr"];
