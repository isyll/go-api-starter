import { PrimaryButton } from "../components/button";
import { EmailLayout } from "../components/layout";
import { FallbackLink, Muted, Paragraph, Title } from "../components/typography";
import { goVar, type Locale } from "../vars";

export type VerifyEmailCopy = {
  preview: string;
  title: string;
  body: string;
  button: string;
  fallback: string;
  ignore: string;
  footer: string;
};

export const verifyEmailCopy: Record<Locale, VerifyEmailCopy> = {
  en: {
    preview: "Confirm your email address",
    title: "Verify your email",
    body: "Thanks for signing up. Click the button below to confirm your email address. This link expires in 24 hours.",
    button: "Verify email",
    fallback: "If the button does not work, copy and paste this link into your browser:",
    ignore: "If you did not create an account, you can safely ignore this email.",
    footer: "You received this email because an account was created with this address.",
  },
  fr: {
    preview: "Confirmez votre adresse email",
    title: "Vérifiez votre email",
    body: "Merci pour votre inscription. Cliquez sur le bouton ci-dessous pour confirmer votre adresse email. Ce lien expire dans 24 heures.",
    button: "Vérifier mon email",
    fallback: "Si le bouton ne fonctionne pas, copiez et collez ce lien dans votre navigateur :",
    ignore: "Si vous n'avez pas créé de compte, vous pouvez ignorer cet email.",
    footer: "Vous recevez cet email car un compte a été créé avec cette adresse.",
  },
};

type Props = {
  copy: VerifyEmailCopy;
  locale: Locale;
};

export default function VerifyEmail({ copy, locale }: Props) {
  const url = goVar("URL");
  return (
    <EmailLayout locale={locale} preview={copy.preview} footer={copy.footer}>
      <Title>{copy.title}</Title>
      <Paragraph>{copy.body}</Paragraph>
      <PrimaryButton href={url}>{copy.button}</PrimaryButton>
      <Muted>{copy.fallback}</Muted>
      <FallbackLink href={url} />
      <Muted>{copy.ignore}</Muted>
    </EmailLayout>
  );
}

VerifyEmail.PreviewProps = {
  copy: verifyEmailCopy.en,
  locale: "en",
} satisfies Props;
