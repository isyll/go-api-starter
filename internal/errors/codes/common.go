package codes

// Cross-cutting codes — HTTP binding, ID decoding, request
// validation, generic 404/405. Bare names: no domain owns them.

var (
	// InternalError — 500. Recovery-middleware fallback when no *HTTPError matched.
	InternalError = register("INTERNAL_ERROR")

	// InvalidPayload — 400. ShouldBindJSON could not decode the request body.
	InvalidPayload = register("INVALID_PAYLOAD")

	// InvalidQuery — 400. URL query parameters failed to bind.
	InvalidQuery = register("INVALID_QUERY")

	// InvalidID — 400. Obfuscated path parameter did not decode to a valid int64.
	InvalidID = register("INVALID_ID")

	// InvalidParam — 400. Required path or query parameter is missing or empty.
	InvalidParam = register("INVALID_PARAM")

	// ValidationError — 400. Per-field validator failures (carries details map).
	ValidationError = register("VALIDATION_ERROR")

	// Forbidden — 403. Policy denied a mutation on a publicly-readable resource.
	// Owner-only resources MUST return a domain-specific 404 to hide existence.
	Forbidden = register("FORBIDDEN")

	// NotFound — 404. Generic resource-type-agnostic miss; prefer domain-specific codes.
	NotFound = register("NOT_FOUND")

	// NoFieldsToUpdate — 400. Update request normalises to a no-op.
	NoFieldsToUpdate = register("NO_FIELDS_TO_UPDATE")

	// NoChanges — 409. State-machine transition targets the current status.
	NoChanges = register("NO_CHANGES")

	// InvalidDateRange — 400. Reversed or otherwise inconsistent date range.
	InvalidDateRange = register("INVALID_DATE_RANGE")

	// DateInPast — 400. Future-only date field received a past timestamp.
	DateInPast = register("DATE_IN_PAST")

	// MethodNotAllowed — 405. Gin NoMethod handler.
	MethodNotAllowed = register("METHOD_NOT_ALLOWED")

	// RouteNotFound — 404. Gin NoRoute handler (routing failure, not resource miss).
	RouteNotFound = register("ROUTE_NOT_FOUND")
)
