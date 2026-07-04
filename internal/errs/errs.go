// Package errs is the gRPC-native domain error model. Domain code returns
// *Error values; the gRPC error interceptor is the only place that turns them
// into a status.Status with rich details (ErrorInfo, BadRequest, LocalizedMessage).
package errs

import (
	"maps"

	grpccodes "google.golang.org/grpc/codes"

	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

// FieldViolation describes one invalid request field.
type FieldViolation struct {
	Field       string
	Description string
}

// Error is a transport-agnostic domain error. It carries the gRPC status code,
// a stable machine-readable app code, an i18n message key, optional field-level
// validation violations, and optional structured data.
type Error struct {
	grpcCode   grpccodes.Code
	appCode    codes.Code
	messageKey string
	fields     []FieldViolation
	data       map[string]any
}

func (e *Error) Error() string                     { return e.messageKey }
func (e *Error) GRPCCode() grpccodes.Code          { return e.grpcCode }
func (e *Error) Code() string                      { return e.appCode.String() }
func (e *Error) MessageKey() string                { return e.messageKey }
func (e *Error) FieldViolations() []FieldViolation { return e.fields }
func (e *Error) Data() map[string]any              { return e.data }

func (e *Error) WithMessage(messageKey string) *Error {
	clone := *e
	clone.messageKey = messageKey
	return &clone
}

func (e *Error) WithCode(code codes.Code) *Error {
	clone := *e
	clone.appCode = code
	return &clone
}

func (e *Error) WithData(data map[string]any) *Error {
	clone := *e
	merged := make(map[string]any, len(e.data)+len(data))
	maps.Copy(merged, e.data)
	maps.Copy(merged, data)
	clone.data = merged
	return &clone
}

func (e *Error) WithFieldViolations(fields ...FieldViolation) *Error {
	clone := *e
	clone.fields = append(append([]FieldViolation(nil), e.fields...), fields...)
	return &clone
}

func newError(grpc grpccodes.Code, code codes.Code, messageKey string) *Error {
	return &Error{grpcCode: grpc, appCode: code, messageKey: messageKey}
}

// New builds an error with an explicit gRPC code.
func New(grpc grpccodes.Code, code codes.Code, messageKey string) *Error {
	return newError(grpc, code, messageKey)
}

func NotFound(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.NotFound, code, messageKey)
}

func BadRequest(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.InvalidArgument, code, messageKey)
}

func Forbidden(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.PermissionDenied, code, messageKey)
}

func Unauthorized(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.Unauthenticated, code, messageKey)
}

func Conflict(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.AlreadyExists, code, messageKey)
}

func Unprocessable(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.FailedPrecondition, code, messageKey)
}

func TooManyRequests(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.ResourceExhausted, code, messageKey)
}

func Internal(code codes.Code, messageKey string) *Error {
	return newError(grpccodes.Internal, code, messageKey)
}

// ServerError is an alias for Internal kept for readability at call sites.
func ServerError(code codes.Code, messageKey string) *Error {
	return Internal(code, messageKey)
}

// Validation builds an InvalidArgument error carrying field-level violations.
func Validation(code codes.Code, messageKey string, fields ...FieldViolation) *Error {
	e := newError(grpccodes.InvalidArgument, code, messageKey)
	e.fields = fields
	return e
}
