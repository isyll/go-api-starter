package interceptor

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/isyll/go-grpc-starter/internal/metrics"
	"github.com/isyll/go-grpc-starter/internal/reqctx"
	idgen "github.com/isyll/go-grpc-starter/pkg/id"
)

// Stream mirrors Unary for streaming RPCs.
func (i *Set) Stream() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		i.recoveryStream,
		i.metricsStream,
		i.requestIDStream,
		i.loggingStream,
		i.localeStream,
		i.errorStream,
		i.authStream,
		i.validationStream,
	}
}

// wrappedStream overrides the stream context to carry request-scoped values.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }

func withStreamContext(ctx context.Context, ss grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{ServerStream: ss, ctx: ctx}
}

func (i *Set) recoveryStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			i.logger.Error("panic in stream handler", "method", info.FullMethod, "panic", r)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(srv, ss)
}

func (i *Set) metricsStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	metrics.GRPCRequestsInFlight.Inc()
	start := time.Now()
	err := handler(srv, ss)
	metrics.GRPCRequestsInFlight.Dec()

	metrics.GRPCRequestsTotal.
		WithLabelValues(info.FullMethod, status.Code(err).String()).
		Inc()
	metrics.GRPCRequestDurationSeconds.
		WithLabelValues(info.FullMethod).
		Observe(time.Since(start).Seconds())
	return err
}

func (i *Set) requestIDStream(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	id := incomingRequestID(ss.Context())
	if id == "" {
		id = idgen.NewUUIDNoDash()
	}
	return handler(srv, withStreamContext(reqctx.WithRequestID(ss.Context(), id), ss))
}

func (i *Set) loggingStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)

	code := status.Code(err)
	fields := []any{
		"method", info.FullMethod,
		"code", code.String(),
		"duration", time.Since(start).String(),
		"request_id", reqctx.RequestIDFromContext(ss.Context()),
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
	}
	if p, ok := peer.FromContext(ss.Context()); ok && p.Addr != nil {
		fields = append(fields, "peer", p.Addr.String())
	}
	if isServerFailure(code) {
		i.logger.Error("grpc", fields...)
	} else {
		i.logger.Info("grpc", fields...)
	}
	return err
}

func (i *Set) localeStream(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := reqctx.WithLanguage(ss.Context(), i.resolveLanguage(ss.Context()))
	return handler(srv, withStreamContext(ctx, ss))
}

func (i *Set) errorStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := handler(srv, ss); err != nil {
		mapped := mapError(ss.Context(), err, i.locale)
		if status.Code(mapped) == codes.Internal {
			i.logger.Error(
				"unhandled error",
				"method", info.FullMethod,
				"error", err.Error(),
				"request_id", reqctx.RequestIDFromContext(ss.Context()),
			)
		}
		return mapped
	}
	return nil
}

func (i *Set) authStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if publicMethods[info.FullMethod] {
		return handler(srv, ss)
	}

	user, session, err := i.authenticate(ss.Context())
	if err != nil {
		return err
	}
	if strings.HasPrefix(info.FullMethod, adminServicePrefix) && !user.IsAdmin() {
		return status.Error(codes.PermissionDenied, "auth.admin_required")
	}

	ctx := reqctx.WithSubject(ss.Context(), reqctx.Subject{
		UserID:    user.ID,
		Role:      reqctx.Role(user.Role),
		SessionID: session.ID,
		DeviceID:  session.DeviceID,
		IsAdmin:   user.IsAdmin(),
	})
	return handler(srv, withStreamContext(ctx, ss))
}

func (i *Set) validationStream(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	return handler(srv, &validatingStream{ServerStream: ss, set: i})
}

type validatingStream struct {
	grpc.ServerStream
	set *Set
}

func (s *validatingStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if msg, ok := m.(proto.Message); ok {
		return s.set.validateMessage(msg)
	}
	return nil
}
