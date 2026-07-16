package interceptor

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/isyll/go-grpc-starter/internal/errs"
	appcodes "github.com/isyll/go-grpc-starter/internal/errs/codes"
)

// Runs innermost, after auth, so rejected requests keep caller identity in logs.
func (i *Set) validationUnary(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if err := i.validateMessage(req); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (i *Set) validateMessage(req any) error {
	if i.validator == nil {
		return nil
	}
	msg, ok := req.(proto.Message)
	if !ok {
		return nil
	}
	err := i.validator.Validate(msg)
	if err == nil {
		return nil
	}
	var ve *protovalidate.ValidationError
	if errors.As(err, &ve) {
		return errs.Validation(
			appcodes.ValidationError,
			"common.validation_error",
			violationsToFields(ve)...,
		)
	}
	// A rule's compile/runtime failure is a server bug, not bad input.
	return errs.Internal(appcodes.InternalError, "common.internal_error")
}

func violationsToFields(ve *protovalidate.ValidationError) []errs.FieldViolation {
	out := make([]errs.FieldViolation, 0, len(ve.Violations))
	for _, v := range ve.Violations {
		out = append(out, errs.FieldViolation{
			Field:       protovalidate.FieldPathString(v.Proto.GetField()),
			Description: v.Proto.GetMessage(),
		})
	}
	return out
}
