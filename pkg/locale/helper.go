package locale

import (
	"context"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Helper provides request-scoped translation helpers.  It wraps the
// per-request *goi18n.Localizer stored in the context.
type Helper struct {
	bundle      *Bundle
	defaultLang string
}

// NewHelper creates a Helper backed by bundle with the given default language.
func NewHelper(bundle *Bundle, defaultLang string) *Helper {
	return &Helper{bundle: bundle, defaultLang: defaultLang}
}

// T translates messageID using the Localizer stored in ctx, applying optional
// template data.  Falls back to messageID when the key is missing.
func (h *Helper) T(
	ctx context.Context,
	id string,
	data ...map[string]any,
) string {
	loc := h.localizerFromCtx(ctx)
	return localize(loc, id, 0, false, data...)
}

// Pluralize translates messageID using CLDR plural rules for count.
func (h *Helper) Pluralize(
	ctx context.Context,
	id string,
	count int,
	data map[string]any,
) string {
	loc := h.localizerFromCtx(ctx)
	return localize(loc, id, count, true, data)
}

// GetLanguage returns the resolved language for the current request.
func (h *Helper) GetLanguage(ctx context.Context) string {
	return GetLanguageFromContext(ctx, h.defaultLang)
}

// GetBundle returns the underlying Bundle (for background workers).
func (h *Helper) GetBundle() *Bundle {
	return h.bundle
}

func (h *Helper) localizerFromCtx(
	ctx context.Context,
) *goi18n.Localizer {
	if loc, ok := ctx.Value(LocalizerContextKey).(*goi18n.Localizer); ok &&
		loc != nil {
		return loc
	}
	if h == nil || h.bundle == nil {
		// No bundle available (e.g. in tests) — return a no-op localizer
		// that always falls back to the message ID.
		return goi18n.NewLocalizer(goi18n.NewBundle(language.English))
	}
	return h.bundle.NewLocalizer(h.defaultLang)
}
