package emails

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/resend/resend-go/v3"

	"github.com/isyll/go-api-starter/internal/metrics"
	"github.com/isyll/go-api-starter/pkg/config"
	"github.com/isyll/go-api-starter/pkg/locale"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// Processor handles email:send and email:scheduled Asynq tasks. It
// renders Go HTML templates with locale-specific data, sends via the
// Resend API, and recovers from panics so a malformed payload cannot
// kill the worker goroutine.
type Processor struct {
	client    *resend.Client
	cfg       *config.EmailConfig
	logger    *logger.Logger
	localizer *locale.Bundle
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewProcessor creates a new email processor
func NewProcessor(
	cfg *config.EmailConfig,
	logx *logger.Logger,
	localizer *locale.Bundle,
) *Processor {
	client := resend.NewClient(cfg.Email.APIKey)

	p := &Processor{
		client:    client,
		cfg:       cfg,
		logger:    logx,
		localizer: localizer,
		templates: make(map[string]*template.Template),
	}

	// Preload templates
	if err := p.loadTemplates(); err != nil {
		logx.Warn("Failed to preload email templates", "error", err)
	}

	return p
}

func (p *Processor) loadTemplates() error {
	basePath := p.cfg.Email.Templates.BasePath

	return filepath.WalkDir(
		basePath,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || filepath.Ext(path) != ".html" {
				return nil
			}

			relPath, _ := filepath.Rel(basePath, path)
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				p.logger.Warn("Failed to parse template",
					"path", path,
					"error", err,
				)
				return nil
			}

			p.mu.Lock()
			p.templates[relPath] = tmpl
			p.mu.Unlock()

			return nil
		},
	)
}

// ProcessTask handles incoming email tasks
func (p *Processor) ProcessTask(
	ctx context.Context,
	t *asynq.Task,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			p.logger.Error(
				"emails worker panic recovered",
				"task_type", t.Type(),
				"panic", r,
				"stack_trace", string(stack),
			)
			metrics.WorkerPanicsTotal.
				WithLabelValues("emails").
				Inc()
			err = fmt.Errorf(
				"emails worker panic: %v", r,
			)
		}
	}()
	switch t.Type() {
	case TaskSendEmail, TaskScheduledEmail:
		var email Email
		if err := json.Unmarshal(t.Payload(), &email); err != nil {
			return fmt.Errorf("failed to unmarshal email: %w", err)
		}
		return p.processEmail(&email)

	case TaskBulkEmail:
		var emails []*Email
		if err := json.Unmarshal(t.Payload(), &emails); err != nil {
			return fmt.Errorf(
				"failed to unmarshal bulk emails: %w",
				err,
			)
		}
		return p.processBulkEmail(emails)

	default:
		return fmt.Errorf("unknown task type: %s", t.Type())
	}
}

func (p *Processor) processEmail(email *Email) error {
	// Get sender based on email type
	senderType := GetSenderType(email.Type)
	sender := GetSender(p.cfg, senderType)

	// Render template if specified
	htmlBody := email.HTMLBody
	if email.TemplateID != "" && htmlBody == "" {
		rendered, err := p.renderTemplate(
			email.TemplateID,
			email.Language,
			email.TemplateData,
		)
		if err != nil {
			p.logger.Warn("Failed to render template, using plain text",
				"template", email.TemplateID,
				"error", err,
			)
		} else {
			htmlBody = rendered
		}
	}

	// Build request
	params := &resend.SendEmailRequest{
		From:    p.formatSender(sender, senderType, email.Language),
		To:      email.To,
		Subject: email.Subject,
		Html:    htmlBody,
		Text:    email.TextBody,
		Tags:    p.buildTags(email),
	}

	if email.ReplyTo != "" {
		params.ReplyTo = email.ReplyTo
	}
	if len(email.CC) > 0 {
		params.Cc = email.CC
	}
	if len(email.BCC) > 0 {
		params.Bcc = email.BCC
	}
	if len(email.Attachments) > 0 {
		params.Attachments = p.buildAttachments(email.Attachments)
	}
	if len(email.Headers) > 0 {
		params.Headers = email.Headers
	}

	// Send email
	sent, err := p.client.Emails.Send(params)
	if err != nil {
		p.logger.Error("Failed to send email",
			"type", email.Type,
			"to", email.To,
			"error", err,
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	p.logger.Info("Email sent successfully",
		"message_id", sent.Id,
		"type", email.Type,
		"to", email.To,
		"sender", sender.Address,
	)

	return nil
}

func (p *Processor) processBulkEmail(emails []*Email) error {
	var lastErr error
	successCount := 0

	for _, email := range emails {
		if err := p.processEmail(email); err != nil {
			lastErr = err
			p.logger.Error("Bulk email failed",
				"type", email.Type,
				"to", email.To,
				"error", err,
			)
		} else {
			successCount++
		}
	}

	p.logger.Info("Bulk email batch completed",
		"total", len(emails),
		"success", successCount,
		"failed", len(emails)-successCount,
	)

	return lastErr
}

func (p *Processor) renderTemplate(
	templateID string,
	lang string,
	data map[string]any,
) (string, error) {
	// Try language-specific template first
	templatePath := filepath.Join(templateID, lang+".html")

	p.mu.RLock()
	tmpl, ok := p.templates[templatePath]
	p.mu.RUnlock()

	if !ok {
		// Fallback to default language
		templatePath = filepath.Join(
			templateID,
			p.cfg.Email.Templates.DefaultLanguage+".html",
		)
		p.mu.RLock()
		tmpl, ok = p.templates[templatePath]
		p.mu.RUnlock()

		if !ok {
			return "", fmt.Errorf("template not found: %s", templateID)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (p *Processor) buildTags(email *Email) []resend.Tag {
	tags := []resend.Tag{
		{Name: "type", Value: string(email.Type)},
		{
			Name:  "sender_category",
			Value: string(GetSenderType(email.Type)),
		},
	}

	for k, v := range email.Tags {
		tags = append(tags, resend.Tag{Name: k, Value: v})
	}

	return tags
}

func (p *Processor) buildAttachments(
	attachments []*Attachment,
) []*resend.Attachment {
	result := make([]*resend.Attachment, len(attachments))
	for i, a := range attachments {
		result[i] = &resend.Attachment{
			Filename:    a.Filename,
			Content:     a.Content,
			ContentType: a.ContentType,
		}
	}
	return result
}

// formatSender returns formatted string
// e.g. "App News <news@app_owner.app>"
func (p *Processor) formatSender(
	s *config.SenderInfo,
	senderType SenderType,
	lang string,
) string {
	name := s.Name
	if lang != "" {
		name = p.localizer.T(lang, "email.fromName."+string(senderType))
	}
	if name == "" {
		return s.Address
	}
	return name + " <" + s.Address + ">"
}
