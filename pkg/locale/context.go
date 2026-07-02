package locale

import "context"

type contextKey string

const (
	LanguageContextKey  contextKey = "language"
	LocalizerContextKey contextKey = "localizer"
)

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
