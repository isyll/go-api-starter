import { Button, Section } from "@react-email/components";
import type { CSSProperties, ReactNode } from "react";
import { tokens } from "../theme/tokens";

type Props = {
  href: string;
  children: ReactNode;
};

export function PrimaryButton({ href, children }: Props) {
  return (
    <Section style={wrapper}>
      <Button href={href} style={button}>
        {children}
      </Button>
    </Section>
  );
}

const wrapper: CSSProperties = {
  textAlign: "center",
  margin: "24px 0",
};

const button: CSSProperties = {
  backgroundColor: tokens.colors.primary,
  borderRadius: tokens.layout.buttonRadius,
  color: tokens.colors.onPrimary,
  fontSize: tokens.typography.baseSize,
  fontWeight: 600,
  padding: "12px 28px",
  textDecoration: "none",
};
