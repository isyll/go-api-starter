package emails

import (
	"time"

	"github.com/isyll/go-api-starter/pkg/config"
)

// Priority controls email processing order.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// EmailType categorizes an email for sender selection.
type EmailType string

const (
	// No-reply / automated emails.
	TypeSystemNotification EmailType = "system.notification"

	// Security emails.
	TypeEmailVerification EmailType = "security.email_verification"
	TypePasswordReset     EmailType = "security.password_reset"
	TypeLoginAlert        EmailType = "security.login_alert"
	TypeAccountChange     EmailType = "security.account_change"

	// Marketing / news emails.
	TypeNewsletter EmailType = "news.newsletter"
	TypePromotion  EmailType = "news.promotion"
)

// SenderType selects the from-identity used to send an email.
type SenderType string

const (
	SenderNoReply  SenderType = "noreply"
	SenderSecurity SenderType = "security"
	SenderNews     SenderType = "news"
)

// Email is a message to be sent by the email worker.
type Email struct {
	Type           EmailType         `json:"type"`
	To             []string          `json:"to"`
	Subject        string            `json:"subject"`
	TemplateID     string            `json:"template_id,omitempty"`
	TemplateData   map[string]any    `json:"template_data,omitempty"`
	HTMLBody       string            `json:"html_body,omitempty"`
	TextBody       string            `json:"text_body,omitempty"`
	Language       string            `json:"language,omitempty"`
	Priority       Priority          `json:"priority,omitempty"`
	ScheduledAt    *time.Time        `json:"scheduled_at,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	ReplyTo        string            `json:"reply_to,omitempty"`
	CC             []string          `json:"cc,omitempty"`
	BCC            []string          `json:"bcc,omitempty"`
	Attachments    []*Attachment     `json:"attachments,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// Attachment is an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
}

// SendResult is the outcome of a send.
type SendResult struct {
	Success      bool   `json:"success"`
	MessageID    string `json:"message_id,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetSenderType returns the sender identity for an email type.
func GetSenderType(emailType EmailType) SenderType {
	switch {
	case isSecurityEmail(emailType):
		return SenderSecurity
	case isNewsEmail(emailType):
		return SenderNews
	default:
		return SenderNoReply
	}
}

// GetSender returns the configured sender info for a sender type.
func GetSender(c *config.EmailConfig, senderType SenderType) *config.SenderInfo {
	senders := c.Email.Senders
	switch senderType {
	case SenderSecurity:
		return &senders.Security
	case SenderNews:
		return &senders.News
	default:
		return &senders.NoReply
	}
}

func isSecurityEmail(t EmailType) bool {
	switch t {
	case TypeEmailVerification, TypePasswordReset, TypeLoginAlert, TypeAccountChange:
		return true
	}
	return false
}

func isNewsEmail(t EmailType) bool {
	switch t {
	case TypeNewsletter, TypePromotion:
		return true
	}
	return false
}
