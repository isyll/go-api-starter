import { PrimaryButton } from "../components/button";
import { EmailLayout } from "../components/layout";
import { FallbackLink, Muted, Paragraph, Title } from "../components/typography";
import { goVar, type Locale } from "../vars";

export type PasswordResetCopy = {
  preview: string;
  title: string;
  body: string;
  button: string;
  fallback: string;
  ignore: string;
  footer: string;
};

export const passwordResetCopy: Record<Locale, PasswordResetCopy> = {
  en: {
    preview: "Reset your password",
    title: "Reset your password",
    body: "We received a request to reset your password. Click the button below to choose a new one. This link expires in 1 hour.",
    button: "Reset password",
    fallback: "If the button does not work, copy and paste this link into your browser:",
    ignore: "If you did not request a password reset, you can safely ignore this email. Your password stays unchanged.",
    footer: "You received this email because a password reset was requested for this address.",
  },
  fr: {
    preview: "Réinitialisez votre mot de passe",
    title: "Réinitialisez votre mot de passe",
    body: "Nous avons reçu une demande de réinitialisation de votre mot de passe. Cliquez sur le bouton ci-dessous pour en choisir un nouveau. Ce lien expire dans 1 heure.",
    button: "Réinitialiser le mot de passe",
    fallback: "Si le bouton ne fonctionne pas, copiez et collez ce lien dans votre navigateur :",
    ignore: "Si vous n'avez pas demandé de réinitialisation, vous pouvez ignorer cet email. Votre mot de passe reste inchangé.",
    footer: "Vous recevez cet email car une réinitialisation de mot de passe a été demandée pour cette adresse.",
  },
};

type Props = {
  copy: PasswordResetCopy;
  locale: Locale;
};

export default function PasswordReset({ copy, locale }: Props) {
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

PasswordReset.PreviewProps = {
  copy: passwordResetCopy.en,
  locale: "en",
} satisfies Props;
