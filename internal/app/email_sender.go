package app

import (
	"context"

	"github.com/isyll/go-api-starter/internal/worker/emails"
)

// emailSender adapts the email worker dispatcher to auth.EmailSender.
type emailSender struct {
	disp   emails.Dispatcher
	webURL string
}

func newEmailSender(disp emails.Dispatcher, webURL string) *emailSender {
	return &emailSender{disp: disp, webURL: webURL}
}

// SendVerificationEmail enqueues an email-verification message.
func (e *emailSender) SendVerificationEmail(ctx context.Context, to, token string) error {
	return e.disp.Send(ctx, &emails.Email{
		Type:    emails.TypeEmailVerification,
		To:      []string{to},
		Subject: "Verify your email",
		TemplateData: map[string]any{
			"token": token,
			"url":   e.webURL + "/verify-email?token=" + token,
		},
	})
}

// SendPasswordResetEmail enqueues a password-reset message.
func (e *emailSender) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	return e.disp.Send(ctx, &emails.Email{
		Type:    emails.TypePasswordReset,
		To:      []string{to},
		Subject: "Reset your password",
		TemplateData: map[string]any{
			"token": token,
			"url":   e.webURL + "/reset-password?token=" + token,
		},
	})
}
