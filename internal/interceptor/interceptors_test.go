package interceptor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authv1 "github.com/isyll/go-grpc-starter/gen/auth/v1"
	"github.com/isyll/go-grpc-starter/internal/auth"
	"github.com/isyll/go-grpc-starter/internal/errs"
	appcodes "github.com/isyll/go-grpc-starter/internal/errs/codes"
	"github.com/isyll/go-grpc-starter/internal/reqctx"
	"github.com/isyll/go-grpc-starter/internal/users"
	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
	apptoken "github.com/isyll/go-grpc-starter/pkg/token"
)

// ---- fakes -----------------------------------------------------------------

type stubTokens struct {
	claims map[string]*apptoken.AccessTokenClaims
}

func (s *stubTokens) Generate(context.Context, int64, int64, string) (string, error) {
	return "", errors.New("not used")
}

func (s *stubTokens) Validate(_ context.Context, token string) (*apptoken.AccessTokenClaims, error) {
	c, ok := s.claims[token]
	if !ok {
		return nil, apptoken.ErrTokenNotFound
	}
	return c, nil
}

func (s *stubTokens) Revoke(context.Context, string) error { return nil }

type stubSessions struct {
	byID map[int64]*auth.DeviceSession
}

func (s *stubSessions) Create(context.Context, *auth.DeviceSession) error {
	return errors.New("not used")
}

func (s *stubSessions) FindByID(_ context.Context, id int64) (*auth.DeviceSession, error) {
	sess, ok := s.byID[id]
	if !ok {
		return nil, errs.ErrSessionNotFound
	}
	return sess, nil
}

func (s *stubSessions) FindByUserAndDeviceID(context.Context, int64, string) (*auth.DeviceSession, error) {
	return nil, errors.New("not used")
}

func (s *stubSessions) Revoke(context.Context, string, int64) (*auth.DeviceSession, error) {
	return nil, errors.New("not used")
}

func (s *stubSessions) FindActiveDevicesByUser(context.Context, int64, time.Duration) ([]auth.DeviceSession, error) {
	return nil, errors.New("not used")
}

func (s *stubSessions) RevokeAllByUserID(context.Context, int64, string) error {
	return errors.New("not used")
}

func (s *stubSessions) RevokeActiveByDeviceID(context.Context, string, string) error {
	return errors.New("not used")
}

func testSet(t *testing.T, tokens *stubTokens, sessions *stubSessions) *Set {
	t.Helper()
	cfg := &config.Configs{Security: &config.SecurityConfig{}}
	cfg.Security.Auth.MaxInactivityTimeout = 24 * time.Hour
	return New(Config{
		Tokens:   tokens,
		Sessions: sessions,
		Cfg:      cfg,
		Logger:   logger.New("test"),
	})
}

func activeSession(userID int64) *auth.DeviceSession {
	return &auth.DeviceSession{
		ID:           7,
		UserID:       userID,
		DeviceID:     "dev-1",
		LastActivity: time.Now().UTC(),
		User: users.User{
			ID:     userID,
			Status: users.UserStatusActive,
			Role:   users.UserRoleUser,
		},
	}
}

func passthroughHandler(called *bool) grpc.UnaryHandler {
	return func(_ context.Context, _ any) (any, error) {
		*called = true
		return "ok", nil
	}
}

// ---- BearerToken -----------------------------------------------------------

func TestBearerToken(t *testing.T) {
	t.Run("missing metadata", func(t *testing.T) {
		_, err := BearerToken(context.Background())
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("missing header", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
		_, err := BearerToken(ctx)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("wrong scheme", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs("authorization", "Basic abc"))
		_, err := BearerToken(ctx)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("valid bearer, case-insensitive scheme", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs("authorization", "bearer tok-123"))
		tok, err := BearerToken(ctx)
		require.NoError(t, err)
		assert.Equal(t, "tok-123", tok)
	})
}

// ---- locale ----------------------------------------------------------------

func TestParseAcceptLanguage(t *testing.T) {
	cases := []struct{ in, want string }{
		{"fr-CA,fr;q=0.9,en;q=0.8", "fr"},
		{"en", "en"},
		{"DE-de", "de"},
		{" pt-BR ", "pt"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, parseAcceptLanguage(c.in), "input %q", c.in)
	}
}

// ---- auth interceptor -------------------------------------------------------

func unaryInfo(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: method}
}

func TestAuthUnaryAllowsPublicMethodWithoutToken(t *testing.T) {
	set := testSet(t, &stubTokens{}, &stubSessions{})
	called := false
	_, err := set.authUnary(context.Background(), nil,
		unaryInfo("/auth.v1.AuthService/Login"), passthroughHandler(&called))
	require.NoError(t, err)
	assert.True(t, called)
}

func TestAuthUnaryRejectsMissingToken(t *testing.T) {
	set := testSet(t, &stubTokens{}, &stubSessions{})
	called := false
	_, err := set.authUnary(context.Background(), nil,
		unaryInfo("/user.v1.UserService/GetMe"), passthroughHandler(&called))
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.False(t, called)
}

func TestAuthUnaryAttachesSubject(t *testing.T) {
	session := activeSession(42)
	set := testSet(t,
		&stubTokens{claims: map[string]*apptoken.AccessTokenClaims{
			"tok-1": {SessionID: session.ID, UserID: 42, DeviceID: "dev-1"},
		}},
		&stubSessions{byID: map[int64]*auth.DeviceSession{session.ID: session}},
	)

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer tok-1"))

	var got reqctx.Subject
	_, err := set.authUnary(ctx, nil, unaryInfo("/user.v1.UserService/GetMe"),
		func(ctx context.Context, _ any) (any, error) {
			got = reqctx.SubjectFrom(ctx)
			return nil, nil
		})
	require.NoError(t, err)
	assert.Equal(t, int64(42), got.UserID)
	assert.Equal(t, session.ID, got.SessionID)
	assert.False(t, got.IsAdmin)
}

func TestAuthUnaryBlocksNonAdminFromAdminService(t *testing.T) {
	session := activeSession(42)
	set := testSet(t,
		&stubTokens{claims: map[string]*apptoken.AccessTokenClaims{
			"tok-1": {SessionID: session.ID, UserID: 42},
		}},
		&stubSessions{byID: map[int64]*auth.DeviceSession{session.ID: session}},
	)

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer tok-1"))
	called := false
	_, err := set.authUnary(ctx, nil,
		unaryInfo("/admin.v1.AdminService/ListUsers"), passthroughHandler(&called))
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
	assert.False(t, called)
}

func TestAuthUnaryRejectsRevokedSession(t *testing.T) {
	session := activeSession(42)
	now := time.Now()
	session.RevokedAt = &now
	set := testSet(t,
		&stubTokens{claims: map[string]*apptoken.AccessTokenClaims{
			"tok-1": {SessionID: session.ID, UserID: 42},
		}},
		&stubSessions{byID: map[int64]*auth.DeviceSession{session.ID: session}},
	)

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer tok-1"))
	called := false
	_, err := set.authUnary(ctx, nil,
		unaryInfo("/user.v1.UserService/GetMe"), passthroughHandler(&called))
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.False(t, called)
}

// ---- error mapping ----------------------------------------------------------

func TestMapErrorBuildsRichStatus(t *testing.T) {
	domainErr := errs.Validation(
		appcodes.ValidationError, "common.validation_error",
		errs.FieldViolation{Field: "email", Description: "must be a valid email"},
	)

	err := mapError(context.Background(), domainErr, nil)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	var gotInfo *errdetails.ErrorInfo
	var gotBad *errdetails.BadRequest
	for _, d := range st.Details() {
		switch v := d.(type) {
		case *errdetails.ErrorInfo:
			gotInfo = v
		case *errdetails.BadRequest:
			gotBad = v
		}
	}
	require.NotNil(t, gotInfo)
	assert.Equal(t, "VALIDATION_ERROR", gotInfo.GetReason())
	require.NotNil(t, gotBad)
	require.Len(t, gotBad.GetFieldViolations(), 1)
	assert.Equal(t, "email", gotBad.GetFieldViolations()[0].GetField())
}

func TestMapErrorHidesUnknownErrors(t *testing.T) {
	err := mapError(context.Background(), errors.New("pq: connection refused"), nil)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Equal(t, "internal error", st.Message())
}

// ---- protovalidate ----------------------------------------------------------

func TestValidateMessageRejectsInvalidRequest(t *testing.T) {
	set := testSet(t, &stubTokens{}, &stubSessions{})

	err := set.validateMessage(&authv1.RegisterRequest{
		Email:    "not-an-email",
		Password: "short",
	})
	require.Error(t, err)

	var domainErr *errs.Error
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, codes.InvalidArgument, domainErr.GRPCCode())
	assert.NotEmpty(t, domainErr.FieldViolations())
}

func TestValidateMessageAcceptsValidRequest(t *testing.T) {
	set := testSet(t, &stubTokens{}, &stubSessions{})

	err := set.validateMessage(&authv1.RegisterRequest{
		Email:     "ada@example.com",
		Password:  "long enough password",
		FirstName: "Ada",
		LastName:  "Lovelace",
	})
	assert.NoError(t, err)
}
