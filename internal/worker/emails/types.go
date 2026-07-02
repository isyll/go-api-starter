package emails

import (
	"time"

	"github.com/isyll/go-grpc-starter/pkg/config"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

type EmailType string

const (
	TypeSystemNotification EmailType = "system.notification"

	TypeEmailVerification EmailType = "security.email_verification"
	TypePasswordReset     EmailType = "security.password_reset"
	TypeLoginAlert        EmailType = "security.login_alert"
	TypeAccountChange     EmailType = "security.account_change"

	TypeNewsletter EmailType = "news.newsletter"
	TypePromotion  EmailType = "news.promotion"
)

type SenderType string

const (
	SenderNoReply  SenderType = "noreply"
	SenderSecurity SenderType = "security"
	SenderNews     SenderType = "news"
)

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

type Attachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
}

type SendResult struct {
	Success      bool   `json:"success"`
	MessageID    string `json:"message_id,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

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
