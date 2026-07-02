package authz

import "context"

type subjectKey struct{}

func WithSubject(ctx context.Context, s Subject) context.Context {
	return context.WithValue(ctx, subjectKey{}, s)
}

func From(ctx context.Context) Subject {
	s, _ := ctx.Value(subjectKey{}).(Subject)
	return s
}

func MustFrom(ctx context.Context) Subject {
	s := From(ctx)
	if s.IsAnonymous() {
		panic("authz: subject required but not set")
	}
	return s
}
