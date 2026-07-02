package locale

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/isyll/go-api-starter/pkg/config"
)

type Bundle struct {
	bundle      *goi18n.Bundle
	defaultLang string
}

func New(cfg *config.AppConfig) (*Bundle, error) {
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	err := filepath.WalkDir(
		cfg.I18n.LocalesDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".toml") {
				return nil
			}
			//nolint:gosec // path is constrained by trusted configured locales directory WalkDir
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf(
					"locale: read %s: %w", path, readErr,
				)
			}
			if _, parseErr := b.ParseMessageFileBytes(
				data, d.Name(),
			); parseErr != nil {
				return fmt.Errorf(
					"locale: parse %s: %w", path, parseErr,
				)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &Bundle{
		bundle:      b,
		defaultLang: cfg.I18n.DefaultLanguage,
	}, nil
}

func (b *Bundle) NewLocalizer(langs ...string) *goi18n.Localizer {
	return goi18n.NewLocalizer(b.bundle, langs...)
}

func (b *Bundle) T(
	lang, id string,
	data ...map[string]any,
) string {
	loc := b.NewLocalizer(lang, b.defaultLang)
	return localize(loc, id, 0, false, data...)
}

func (b *Bundle) SupportedLanguages() []string {
	tags := b.bundle.LanguageTags()
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = t.String()
	}
	return out
}

func (b *Bundle) DefaultLanguage() string {
	return b.defaultLang
}

func localize(
	loc *goi18n.Localizer,
	id string,
	pluralCount int,
	usePlural bool,
	data ...map[string]any,
) string {
	cfg := &goi18n.LocalizeConfig{MessageID: id}

	if len(data) > 0 && data[0] != nil {
		cfg.TemplateData = data[0]
	}

	if usePlural {
		cfg.PluralCount = pluralCount
		if cfg.TemplateData == nil {
			cfg.TemplateData = map[string]any{"count": pluralCount}
		} else {
			cfg.TemplateData.(map[string]any)["count"] = pluralCount
		}
	}

	msg, err := loc.Localize(cfg)
	if err != nil || msg == "" {
		return id
	}
	return msg
}
