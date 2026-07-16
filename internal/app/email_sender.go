package app

import (
	"context"
	"time"

	"github.com/isyll/go-grpc-starter/internal/reqctx"
	"github.com/isyll/go-grpc-starter/internal/worker/emails"
	"github.com/isyll/go-grpc-starter/pkg/locale"
)

type emailSender struct {
	disp   emails.Dispatcher
	webURL string
	bundle *locale.Bundle
}

func newEmailSender(disp emails.Dispatcher, webURL string, bundle *locale.Bundle) *emailSender {
	return &emailSender{disp: disp, webURL: webURL, bundle: bundle}
}

func (e *emailSender) SendVerificationEmail(ctx context.Context, to, token string) error {
	lang := e.language(ctx)
	return e.disp.Send(ctx, &emails.Email{
		Type:       emails.TypeEmailVerification,
		To:         []string{to},
		Subject:    e.subject(lang, "email.verify.subject", "Verify your email"),
		TemplateID: emails.TemplateVerifyEmail,
		Language:   lang,
		TemplateData: templateData(
			e.webURL+"/verify-email?token="+token, token,
		),
	})
}

func (e *emailSender) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	lang := e.language(ctx)
	return e.disp.Send(ctx, &emails.Email{
		Type:       emails.TypePasswordReset,
		To:         []string{to},
		Subject:    e.subject(lang, "email.reset.subject", "Reset your password"),
		TemplateID: emails.TemplatePasswordReset,
		Language:   lang,
		TemplateData: templateData(
			e.webURL+"/reset-password?token="+token, token,
		),
	})
}

func (e *emailSender) language(ctx context.Context) string {
	if lang := reqctx.LanguageFromContext(ctx); lang != "" {
		return lang
	}
	if e.bundle != nil {
		return e.bundle.DefaultLanguage()
	}
	return "en"
}

func (e *emailSender) subject(lang, key, fallback string) string {
	if e.bundle == nil {
		return fallback
	}
	if s := e.bundle.T(lang, key); s != "" && s != key {
		return s
	}
	return fallback
}

// Keys must match the generated template placeholders.
func templateData(url, token string) map[string]any {
	return map[string]any{
		"URL":   url,
		"Token": token,
		"Year":  time.Now().UTC().Year(),
	}
}
