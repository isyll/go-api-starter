package authz

import "context"

type subjectKey struct{}

// WithSubject stores s in ctx and returns the new context.
func WithSubject(ctx context.Context, s Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

// From retrieves the Subject from ctx.
// Returns a zero-value anonymous Subject when none is set.
func From(ctx context.Context) Subject {
	s, _ := ctx.Value(subjectKey{}).(Subject)
	return s
}

// MustFrom retrieves the Subject and panics when it has not been
// set. Use only in handlers protected by AuthRequired.
func MustFrom(ctx context.Context) Subject {
	s := From(ctx)
	if s.IsAnonymous() {
		panic("authz: subject required but not set")
	}
	return s
}
