package grpc

import (
	"context"
	"strings"
	"time"

	"github.com/isyll/go-grpc-starter/internal/authz"
	"github.com/isyll/go-grpc-starter/internal/domain/auth"
	"github.com/isyll/go-grpc-starter/internal/models"
	"github.com/isyll/go-grpc-starter/internal/reqctx"
	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
	apptoken "github.com/isyll/go-grpc-starter/pkg/token"
	"github.com/isyll/go-grpc-starter/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/api.v1.HealthService/Check":              true,
	"/api.v1.HealthService/Ready":              true,
	"/api.v1.AuthService/Register":             true,
	"/api.v1.AuthService/Login":                true,
	"/api.v1.AuthService/RefreshToken":         true,
	"/api.v1.AuthService/VerifyEmail":          true,
	"/api.v1.AuthService/RequestPasswordReset": true,
	"/api.v1.AuthService/ResetPassword":        true,
}

const adminServicePrefix = "/api.v1.AdminService/"

type interceptors struct {
	tokens   apptoken.AccessTokenManager
	sessions auth.DeviceSessionRepository
	cfg      *config.Configs
	logger   *logger.Logger
}

func (i *interceptors) authUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if publicMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	user, session, err := i.authenticate(ctx)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(info.FullMethod, adminServicePrefix) && !user.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "auth.admin_required")
	}

	ctx = withUser(ctx, user)
	ctx = withSessionID(ctx, session.ID)
	ctx = authz.WithSubject(ctx, authz.Subject{
		UserID:    user.ID,
		Role:      authz.Role(user.Role),
		SessionID: session.ID,
		DeviceID:  session.DeviceID,
		IsAdmin:   user.IsAdmin(),
	})
	return handler(ctx, req)
}

func (i *interceptors) authenticate(ctx context.Context) (*models.User, *models.DeviceSession, error) {
	token, err := bearerToken(ctx)
	if err != nil {
		return nil, nil, err
	}
	claims, err := i.tokens.Validate(ctx, token)
	if err != nil {
		return nil, nil, status.Error(codes.Unauthenticated, "auth.invalid_token")
	}
	session, err := i.sessions.FindByID(ctx, claims.SessionID)
	if err != nil || session.IsRevoked() {
		return nil, nil, status.Error(codes.Unauthenticated, "auth.session_invalid")
	}
	if session.IsInactive(i.cfg.Security.Auth.MaxInactivityTimeout) {
		return nil, nil, status.Error(codes.Unauthenticated, "auth.session_expired")
	}
	if !session.User.IsActive() {
		return nil, nil, status.Error(codes.PermissionDenied, "auth.account_inactive")
	}
	return &session.User, session, nil
}

func bearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "auth.missing_metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "auth.missing_token")
	}
	parts := strings.SplitN(values[0], " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", status.Error(codes.Unauthenticated, "auth.invalid_token_format")
	}
	return parts[1], nil
}

func (i *interceptors) recoveryUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			i.logger.Error("panic in handler", "method", info.FullMethod, "panic", r)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}

func (i *interceptors) requestIDUnary(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	id := incomingRequestID(ctx)
	if id == "" {
		id = utils.NewUUIDNoDash()
	}
	return handler(reqctx.WithRequestID(ctx, id), req)
}

func (i *interceptors) loggingUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	i.logger.Info(
		"grpc",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start).String(),
	)
	return resp, err
}

func incomingRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if v := md.Get("x-request-id"); len(v) > 0 {
		return v[0]
	}
	return ""
}
