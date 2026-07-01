package codes

// Suspension-subdomain codes. User-suspension state is owned by
// the users domain; this code surfaces from the dedicated
// suspension helper.
//
// [SuspensionNotFound] is intentionally distinct from
// [AccountNotSuspended] (declared in users.go): the latter is
// returned by the user-domain lookup, the former by the
// dedicated suspension helper. Keeping the two means clients can
// tell which surface answered.

// SuspensionNotFound — 404. Returned by the suspension
// helper when no active suspension is found for the user.
var SuspensionNotFound = register("SUSPENSION_NOT_FOUND")
