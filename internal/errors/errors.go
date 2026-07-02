// Package apperrors defines the canonical error type and sentinels.
package apperrors

import (
	"maps"
	"net/http"

	"github.com/isyll/go-api-starter/internal/errors/codes"
)

type HTTPError struct {
	status     int
	code       codes.Code
	messageKey string
	details    map[string]string
	data       map[string]any
}

func (e *HTTPError) Error() string {
	return e.messageKey
}

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

func (e *HTTPError) WithDetails(
	details map[string]string,
) *HTTPError {
	clone := *e
	clone.details = details
	return &clone
}

func (e *HTTPError) WithMessage(
	messageKey string,
) *HTTPError {
	clone := *e
	clone.messageKey = messageKey
	return &clone
}

func (e *HTTPError) WithCode(code codes.Code) *HTTPError {
	clone := *e
	clone.code = code
	return &clone
}

func (e *HTTPError) Status() int { return e.status }

func (e *HTTPError) Code() string { return e.code.String() }

func (e *HTTPError) MessageKey() string {
	return e.messageKey
}

func (e *HTTPError) Details() map[string]string {
	return e.details
}

func (e *HTTPError) ExtraData() map[string]any {
	return e.data
}

func (e *HTTPError) ToResponse() map[string]any {
	return map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    e.code.String(),
			"message": e.messageKey,
		},
	}
}

func New(
	status int, code codes.Code, messageKey string,
) *HTTPError {
	return &HTTPError{
		status:     status,
		code:       code,
		messageKey: messageKey,
	}
}

func NotFound(code codes.Code, messageKey string) *HTTPError {
	return New(http.StatusNotFound, code, messageKey)
}

func BadRequest(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusBadRequest, code, messageKey)
}

func Forbidden(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusForbidden, code, messageKey)
}

func Unauthorized(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusUnauthorized, code, messageKey,
	)
}

func Conflict(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(http.StatusConflict, code, messageKey)
}

func Unprocessable(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusUnprocessableEntity,
		code,
		messageKey,
	)
}

func TooManyRequests(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusTooManyRequests, code, messageKey,
	)
}

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

func ServerError(
	code codes.Code, messageKey string,
) *HTTPError {
	return New(
		http.StatusInternalServerError,
		code,
		messageKey,
	)
}

func Internal(
	code codes.Code, messageKey string,
) *HTTPError {
	return ServerError(code, messageKey)
}

func (e *HTTPError) Data() map[string]any { return e.data }
