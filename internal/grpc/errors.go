package grpc

import (
	"errors"
	"net/http"

	apperrors "github.com/isyll/go-api-starter/internal/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toStatus converts a domain error into a gRPC status error. Known
// *HTTPError values map to the matching gRPC code and carry their
// i18n message key; everything else becomes Internal.
func toStatus(err error) error {
	if err == nil {
		return nil
	}
	var he *apperrors.HTTPError
	if errors.As(err, &he) {
		return status.Error(httpToGRPC(he.Status()), he.MessageKey())
	}
	return status.Error(codes.Internal, "internal error")
}

func httpToGRPC(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	default:
		return codes.Internal
	}
}
