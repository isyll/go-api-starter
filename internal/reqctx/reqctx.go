// Package reqctx carries request-scoped values across context boundaries.
package reqctx

import "context"

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	languageKey  ctxKey = "language"
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithLanguage stores the request's resolved language tag (e.g. "en").
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageKey, lang)
}

// LanguageFromContext returns the resolved language, or "" if unset.
func LanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(languageKey).(string); ok {
		return lang
	}
	return ""
}
