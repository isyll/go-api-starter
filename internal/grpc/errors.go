package grpc

import (
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"

	"github.com/isyll/go-grpc-starter/internal/errs"
)

const errorDomain = "go-grpc-starter"

// toStatus is the single place that turns a domain *errs.Error into a gRPC
// status with rich details: an ErrorInfo carrying the stable app code and a
// BadRequest carrying any field-level validation violations. Anything that is
// not an *errs.Error is reported as an opaque internal error.
func toStatus(err error) error {
	if err == nil {
		return nil
	}

	var e *errs.Error
	if !errors.As(err, &e) {
		return status.Error(codes.Internal, "internal error")
	}

	st := status.New(e.GRPCCode(), e.MessageKey())

	details := []protoadapt.MessageV1{
		&errdetails.ErrorInfo{Reason: e.Code(), Domain: errorDomain},
	}
	if fvs := e.FieldViolations(); len(fvs) > 0 {
		br := &errdetails.BadRequest{}
		for _, fv := range fvs {
			br.FieldViolations = append(br.FieldViolations,
				&errdetails.BadRequest_FieldViolation{
					Field:       fv.Field,
					Description: fv.Description,
				})
		}
		details = append(details, br)
	}

	if enriched, werr := st.WithDetails(details...); werr == nil {
		st = enriched
	}
	return st.Err()
}
