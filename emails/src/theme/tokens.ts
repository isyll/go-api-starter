export const tokens = {
  brand: {
    name: "App",
    logoUrl: "https://placehold.co/140x36/1D4ED8/FFFFFF?text=APP",
    logoWidth: 140,
    logoHeight: 36,
  },
  colors: {
    primary: "#1D4ED8",
    primaryAccent: "#2563EB",
    onPrimary: "#FFFFFF",
    background: "#F1F5F9",
    surface: "#FFFFFF",
    border: "#E2E8F0",
    text: "#0F172A",
    muted: "#64748B",
    linkText: "#1D4ED8",
  },
  typography: {
    fontFamily: "'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
    baseSize: 15,
    headingSize: 22,
    smallSize: 13,
    lineHeight: "24px",
  },
  layout: {
    maxWidth: 600,
    radius: 14,
    contentPadding: 32,
    headerPadding: 24,
    buttonRadius: 8,
  },
} as const;

export type Tokens = typeof tokens;

export const headerGradient =
  `linear-gradient(135deg, ${tokens.colors.primary} 0%, ${tokens.colors.primaryAccent} 100%)`;
