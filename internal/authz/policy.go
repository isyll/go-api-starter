// Package authz defines the canonical authorization primitives used
// by every owner-scoped route in the App backend.
//
// The package exports three concepts:
//
//   - Subject: the authenticated principal for the current request,
//     stored in the request context by AuthRequired middleware and
//     retrieved via From / MustFrom.
//   - Action: a named operation (read, write, cancel, ...) the
//     Subject wishes to perform on a resource.
//   - Policy[T]: a pure decision function returning nil on allow or
//     a *errors.HTTPError on deny.
//
// Policy implementations live in each domain's policy.go file and
// are invoked by middleware.RequireOwned[T] after the resource is
// pre-loaded from the repository.
//
// Authorization runs as three independent defense layers, all of
// which must be in place for every owner-scoped resource:
//
//  1. Middleware (RequireOwned[T] in internal/middleware) pre-loads
//     the resource via the domain's FindByID(ctx, int64), runs the
//     Policy, stashes the loaded *models.X under a typed ResourceKey,
//     and aborts the chain on any failure.
//  2. Policy (this package) is a pure decision function. No DB
//     access. Decides allow/deny per (Subject, Action, resource)
//     and returns the correct masked error.
//  3. PostgreSQL row-level security policies on every owner-scoped
//     table. Reads filter by app.current_user_id; mutations fail on
//     rows the caller cannot see. The runtime DB role never bypasses
//     RLS; only the migration role has DDL privileges.
//
// No layer trusts another.
package authz

// Action is a named operation a Subject may perform on a resource.
// Action values are passed to Policy.Can to discriminate per-action
// authorization branches (e.g. read vs. cancel vs. start).
type Action string

const (
	// ActRead authorizes loading a single resource by ID.
	ActRead Action = "read"
	// ActList authorizes listing a collection scoped to the caller.
	ActList Action = "list"
	// ActCreate authorizes creating a new resource owned by the caller.
	ActCreate Action = "create"
	// ActWrite authorizes mutating an existing resource (update/edit).
	ActWrite Action = "write"
	// ActDelete authorizes removing (soft-delete) an existing resource.
	ActDelete Action = "delete"
	// ActCancel authorizes the cancel lifecycle transition.
	ActCancel Action = "cancel"
	// ActStart authorizes the start lifecycle transition (driver-only
	// on trips, for example).
	ActStart Action = "start"
	// ActComplete authorizes the complete lifecycle transition.
	ActComplete Action = "complete"
	// ActArchive authorizes archiving (and the inverse unarchive)
	// a resource.
	ActArchive Action = "archive"
)

// Policy decides whether Subject s may perform action on resource.
// Implementations return nil when allowed or a *errors.HTTPError
// when the action is denied.
//
// Policy.Can must be a pure function: no DB access, no side effects.
// The middleware layer is responsible for loading the resource
// before invoking the policy.
//
// Denial conventions:
//
//   - Admin always allows: the first line of Can must be
//     "if s.IsAdmin { return nil }".
//   - Owner-only resources return the domain's 404 not-found
//     sentinel to non-owners. Never confirm existence to a caller
//     who lacks read access.
//   - Resources with a public read surface (e.g. trips with
//     status='available') return 404 on reads but
//     apperrors.ErrForbidden on non-owner mutations
//     (cancel/start/complete/...).
//   - Multi-actor resources (e.g. bookings with both passenger
//     and driver) split per-action inside Can, returning the
//     appropriate sentinel for each branch.
type Policy[T any] interface {
	// Can returns nil when s may perform action on resource, or a
	// *errors.HTTPError describing the denial otherwise.
	Can(s Subject, action Action, resource T) error
}
