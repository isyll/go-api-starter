// Package persistence provides the Unit-of-Work Manager that binds
// a gorm.DB transaction to a context.Context, the Q helper that
// resolves the active transaction or falls back to a plain DB
// session, and the per-request query counter used for observability.
//
// Repositories call Q(ctx, r.db) in place of r.db directly so that
// all writes inside a Manager.WithTx block participate in the same
// transaction, including outbox writes.
package persistence
