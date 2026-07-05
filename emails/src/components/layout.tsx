import {
  Body,
  Container,
  Head,
  Hr,
  Html,
  Img,
  Preview,
  Section,
  Text,
} from "@react-email/components";
import type { CSSProperties, ReactNode } from "react";
import { headerGradient, tokens } from "../theme/tokens";
import { goVar, type Locale } from "../vars";

type Props = {
  locale: Locale;
  preview: string;
  footer: string;
  children: ReactNode;
};

export function EmailLayout({ locale, preview, footer, children }: Props) {
  return (
    <Html lang={locale} dir="ltr">
      <Head />
      <Preview>{preview}</Preview>
      <Body style={body}>
        <Container style={card}>
          <Section style={header}>
            <Img
              src={tokens.brand.logoUrl}
              alt={tokens.brand.name}
              width={tokens.brand.logoWidth}
              height={tokens.brand.logoHeight}
              style={logo}
            />
          </Section>
          <Section style={content}>{children}</Section>
          <Hr style={divider} />
          <Section style={footerSection}>
            <Text style={footerText}>{footer}</Text>
            <Text style={footerText}>
              &copy; {goVar("Year")} {tokens.brand.name}
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
}

const body: CSSProperties = {
  backgroundColor: tokens.colors.background,
  fontFamily: tokens.typography.fontFamily,
  margin: 0,
  padding: 0,
};

const card: CSSProperties = {
  maxWidth: tokens.layout.maxWidth,
  margin: "32px auto",
  backgroundColor: tokens.colors.surface,
  borderRadius: tokens.layout.radius,
  border: `1px solid ${tokens.colors.border}`,
  overflow: "hidden",
};

const header: CSSProperties = {
  background: headerGradient,
  padding: tokens.layout.headerPadding,
  textAlign: "center",
};

const logo: CSSProperties = {
  margin: "0 auto",
  display: "block",
};

const content: CSSProperties = {
  padding: tokens.layout.contentPadding,
};

const divider: CSSProperties = {
  borderColor: tokens.colors.border,
  margin: 0,
};

const footerSection: CSSProperties = {
  padding: `16px ${tokens.layout.contentPadding}px`,
};

const footerText: CSSProperties = {
  color: tokens.colors.muted,
  fontSize: tokens.typography.smallSize,
  lineHeight: "20px",
  margin: "4px 0",
};
