// Package apperrors defines the canonical error type and sentinel
// values used by the entire backend.
//
// The error-handling contract is asymmetric across layers because
// services run from two execution contexts with different recovery
// capabilities — the HTTP request path (panics caught by the
// recovery middleware) and Asynq worker goroutines (panics crash
// the goroutine). The rules:
//
//   - Repositories return *HTTPError for expected failures
//     (record-not-found, constraint violation) and panic for
//     infrastructure failures (connection drop, write failure,
//     corruption). Never return raw error from a DB call.
//   - Services never panic. They return domain sentinel errors
//     (declared as *HTTPError values in each domain's errors.go)
//     or propagate a repository's *HTTPError unchanged. An
//     infrastructure failure inside a service (crypto, cache) is
//     fmt.Errorf-wrapped, never panicked.
//   - Handlers wrapped by httpx.Wrap return error. HandleError
//     detects *HTTPError via errors.As and writes the correct
//     status code + JSON envelope. A handler must never call
//     c.JSON for errors.
//
// HTTPError instances declared as package-level sentinels MUST be
// treated as immutable. The WithData / WithDetails / WithMessage /
// WithCode modifiers all return a clone — they never mutate the
// receiver. Mutating a sentinel would corrupt every future caller.
//
// # Error codes
//
// Every machine-readable error code is a [codes.Code] value
// declared in the internal/errors/codes package. The constructors
// below take [codes.Code], not [string] — the compiler rejects raw
// string literals at the call site, so a code can only ever come
// from the centralized registry. See the codes package doc for the
// naming rule, uniqueness rule, and selection rule.
package apperrors

import (
	"maps"
	"net/http"

	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// HTTPError carries the HTTP-shaped failure information needed to
// produce a canonical API error envelope. Match an arbitrary error
// to an *HTTPError via errors.As (HandleError does this for you).
type HTTPError struct {
	// status is the HTTP status code returned to the client.
	status int
	// code is the machine-readable error code emitted in the
	// response envelope. Always a value from the codes package;
	// the compiler forbids string literals here.
	code codes.Code
	// messageKey is the i18n key (domain.reason_in_snake_case)
	// resolved by the response writer using the request locale.
	messageKey string
	// details carries field-level validation errors keyed by JSON
	// field name; values are i18n keys.
	details map[string]string
	// data carries template variables for the i18n message.
	data map[string]any
}

// Error returns the i18n message key so HTTPError satisfies the
// standard error interface. Higher layers should rely on Status,
// Code, MessageKey, Details, and ExtraData rather than this raw
// representation.
func (e *HTTPError) Error() string {
	return e.messageKey
}

// WithData returns a copy with additional i18n template data merged in.
func (e *HTTPError) WithData(
	data map[string]any,
) *HTTPError {
	clone := *e
	merged := make(
		map[string]any,
		len(e.data)+len(data),
	)
	maps.Copy(merged, e.data)
	maps.Copy(merged, data)
	clone.data = merged
	return &clone
}

// WithDetails returns a copy with field-level validation details.
func (e *HTTPError) WithDetails(
	details map[string]string,
) *HTTPError {
	clone := *e
	clone.details = details
	return &clone
}

// WithMessage returns a copy with a different i18n message key.
func (e *HTTPError) WithMessage(
	messageKey string,
) *HTTPError {
	clone := *e
	clone.messageKey = messageKey
	return &clone
}

// WithCode returns a copy with a different machine-readable error
// code. The replacement MUST be a registered [codes.Code] — passing
// a raw string is rejected at compile time.
//
// Reserved for dynamic dispatch from a single sentinel to one of
// several siblings (e.g. auth resend can resolve to either
// [codes.OtpResendLimitExceeded] or
// [codes.OtpResendTooSoon] depending on rate-limit state).
// Outside this dispatch pattern, prefer declaring a dedicated
// sentinel rather than rewriting the code at the call site.
func (e *HTTPError) WithCode(code codes.Code) *HTTPError {
	clone := *e
	clone.code = code
	return &clone
}

// Status returns the HTTP status code.
func (e *HTTPError) Status() int { return e.status }

// Code returns the wire-format string of the error's [codes.Code].
// The JSON envelope writer uses this on the way out.
func (e *HTTPError) Code() string { return e.code.String() }

// MessageKey returns the i18n message key.
func (e *HTTPError) MessageKey() string {
	return e.messageKey
}

// Details returns field-level validation errors.
func (e *HTTPError) Details() map[string]string {
	return e.details
}

// ExtraData returns the i18n template data map.
func (e *HTTPError) ExtraData() map[string]any {
	return e.data
}

// ToResponse returns the standard API error envelope as a plain map.
// Use in middleware that has no i18n helper — the message field
// carries the raw i18n key, which the admin frontend can resolve
// independently.
func (e *HTTPError) ToResponse() map[string]any {
	return map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    e.code.String(),
			"message": e.messageKey,
		},
	}
}

// New constructs an HTTPError with arbitrary status, code, and
// i18n message key. Prefer the named constructors below.
func New(
	status int, code codes.Code, messageKey string,
) *HTTPError {
	return &HTTPError{
		status:     status,
		code:       code,
		messageKey: messageKey,
	}
}

// NotFound builds a 404 HTTPError. Use for owner-only resources so
// existence is hidden from non-owners (V2 §5.4 masking).
func NotFound(code codes.Code, messageKey string) *HTTPError {
	return New(http.StatusNotFound, code, messageKey)
}

// BadRequest builds a 400 HTTPError for malformed inputs or
// rejected business rules.
func BadRequest(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusBadRequest, code, messageKey)
}

// Forbidden builds a 403 HTTPError. Use for publicly-readable
// resources when the caller is not the owner and is attempting a
// mutation (cancel/start/complete/...).
func Forbidden(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusForbidden, code, messageKey)
}

// Unauthorized builds a 401 HTTPError when the caller is not
// authenticated.
func Unauthorized(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusUnauthorized, code, messageKey,
	)
}

// Conflict builds a 409 HTTPError for state-conflict failures
// (duplicate, invalid status transition, ...).
func Conflict(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusConflict, code, messageKey)
}

// Unprocessable builds a 422 HTTPError for semantically-invalid
// requests that parsed cleanly.
func Unprocessable(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusUnprocessableEntity,
		code,
		messageKey,
	)
}

// TooManyRequests builds a 429 HTTPError for rate-limited callers.
func TooManyRequests(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusTooManyRequests, code, messageKey,
	)
}

// Validation builds a 400 HTTPError carrying per-field validation
// details. Each details entry maps a JSON field name to an i18n
// key. Use with the canonical "common.validation_error" message
// key (and typically codes.ValidationError) when surfacing
// multiple field-level failures.
func Validation(
	code codes.Code, messageKey string,
	details map[string]string,
) *HTTPError {
	return &HTTPError{
		status:     http.StatusBadRequest,
		code:       code,
		messageKey: messageKey,
		details:    details,
	}
}

// ServerError builds a 500 HTTPError. Prefer panicking from a
// repository when the cause is infrastructure failure — the
// recovery middleware will produce the canonical envelope.
func ServerError(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusInternalServerError,
		code,
		messageKey,
	)
}

// Internal is an alias for ServerError for compatibility.
func Internal(
	code codes.Code, messageKey string,
) *HTTPError {
	return ServerError(code, messageKey)
}

// Data returns the i18n template data map (alias of ExtraData).
func (e *HTTPError) Data() map[string]any { return e.data }
