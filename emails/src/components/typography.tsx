import { Heading, Link, Text } from "@react-email/components";
import type { CSSProperties, ReactNode } from "react";
import { tokens } from "../theme/tokens";

export function Title({ children }: { children: ReactNode }) {
  return <Heading style={title}>{children}</Heading>;
}

export function Paragraph({ children }: { children: ReactNode }) {
  return <Text style={paragraph}>{children}</Text>;
}

export function Muted({ children }: { children: ReactNode }) {
  return <Text style={muted}>{children}</Text>;
}

export function FallbackLink({ href }: { href: string }) {
  return (
    <Text style={muted}>
      <Link href={href} style={link}>
        {href}
      </Link>
    </Text>
  );
}

const title: CSSProperties = {
  color: tokens.colors.text,
  fontSize: tokens.typography.headingSize,
  fontWeight: 700,
  margin: "0 0 16px",
};

const paragraph: CSSProperties = {
  color: tokens.colors.text,
  fontSize: tokens.typography.baseSize,
  lineHeight: tokens.typography.lineHeight,
  margin: "0 0 16px",
};

const muted: CSSProperties = {
  color: tokens.colors.muted,
  fontSize: tokens.typography.smallSize,
  lineHeight: "20px",
  margin: "0 0 12px",
  wordBreak: "break-all",
};

const link: CSSProperties = {
  color: tokens.colors.linkText,
  textDecoration: "underline",
};
