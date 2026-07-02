package locale

import (
	"context"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Helper struct {
	bundle      *Bundle
	defaultLang string
}

func NewHelper(bundle *Bundle, defaultLang string) *Helper {
	return &Helper{bundle: bundle, defaultLang: defaultLang}
}

func (h *Helper) T(
	ctx context.Context,
	id string,
	data ...map[string]any,
) string {
	loc := h.localizerFromCtx(ctx)
	return localize(loc, id, 0, false, data...)
}

func (h *Helper) Pluralize(
	ctx context.Context,
	id string,
	count int,
	data map[string]any,
) string {
	loc := h.localizerFromCtx(ctx)
	return localize(loc, id, count, true, data)
}

func (h *Helper) GetLanguage(ctx context.Context) string {
	return GetLanguageFromContext(ctx, h.defaultLang)
}

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
		return goi18n.NewLocalizer(goi18n.NewBundle(language.English))
	}
	return h.bundle.NewLocalizer(h.defaultLang)
}
