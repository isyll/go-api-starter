package emails

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

func testProcessor(t *testing.T) *Processor {
	t.Helper()
	cfg := &config.EmailConfig{}
	cfg.Email.Templates.BasePath = "../../../templates/emails"
	cfg.Email.Templates.DefaultLanguage = "en"
	return NewProcessor(cfg, logger.New("test"), nil)
}

func TestGeneratedTemplatesRenderAllLanguages(t *testing.T) {
	p := testProcessor(t)
	data := map[string]any{
		"URL":   "https://app.example.com/verify-email?token=abc123",
		"Token": "abc123",
		"Year":  2026,
	}

	for _, name := range []string{TemplateVerifyEmail, TemplatePasswordReset} {
		for _, lang := range []string{"en", "fr"} {
			html, err := p.renderTemplate(name, lang, data)
			require.NoError(t, err, "%s/%s", name, lang)
			assert.Contains(t, html, "https://app.example.com/verify-email?token=abc123")
			assert.Contains(t, html, "2026")
			assert.NotContains(t, html, "{{", "unsubstituted placeholder in %s/%s", name, lang)
		}
	}
}

func TestRenderTemplateFallsBackToDefaultLanguage(t *testing.T) {
	p := testProcessor(t)
	html, err := p.renderTemplate(TemplateVerifyEmail, "de", map[string]any{
		"URL": "https://x", "Token": "t", "Year": 2026,
	})
	require.NoError(t, err)
	assert.True(t, strings.Contains(html, `lang="en"`))
}

func TestRenderTemplateUnknownTemplate(t *testing.T) {
	p := testProcessor(t)
	_, err := p.renderTemplate("does_not_exist", "en", nil)
	assert.Error(t, err)
}
