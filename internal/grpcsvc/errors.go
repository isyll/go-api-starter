package grpcsvc

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"

	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/reqctx"
)

const errorDomain = "go-grpc-starter"

// translator resolves an i18n message key against a language. *locale.Bundle
// satisfies it.
type translator interface {
	T(lang, id string, data ...map[string]any) string
	DefaultLanguage() string
}

// mapError is the single place that turns a domain *errs.Error into a gRPC
// status with rich details: an ErrorInfo (stable app code), a LocalizedMessage
// (the message resolved against the request locale), and a BadRequest carrying
// any field-level validation violations. Errors that are already a gRPC status
// pass through; anything else becomes an opaque internal error.
func mapError(ctx context.Context, err error, tr translator) error {
	if err == nil {
		return nil
	}

	var e *errs.Error
	if !errors.As(err, &e) {
		if _, ok := status.FromError(err); ok {
			return err
		}
		return status.Error(codes.Internal, "internal error")
	}

	lang := reqctx.LanguageFromContext(ctx)
	if lang == "" && tr != nil {
		lang = tr.DefaultLanguage()
	}
	message := e.MessageKey()
	if tr != nil {
		message = tr.T(lang, e.MessageKey(), e.Data())
	}

	st := status.New(e.GRPCCode(), message)
	details := []protoadapt.MessageV1{
		&errdetails.ErrorInfo{Reason: e.Code(), Domain: errorDomain},
		&errdetails.LocalizedMessage{Locale: lang, Message: message},
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
