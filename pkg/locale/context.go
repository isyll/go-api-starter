package locale

import "context"

type contextKey string

const (
	// LanguageContextKey stores the resolved language tag string.
	LanguageContextKey contextKey = "language"
	// LocalizerContextKey stores the per-request *goi18n.Localizer.
	LocalizerContextKey contextKey = "localizer"
)

// GetLanguageFromContext returns the request language or defaultLocale.
func GetLanguageFromContext(
	ctx context.Context,
	defaultLocale string,
) string {
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok &&
		lang != "" {
		return lang
	}
	return defaultLocale
}
