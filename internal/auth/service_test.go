package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isyll/go-grpc-starter/internal/event"
	"github.com/isyll/go-grpc-starter/internal/platform/cache"
	"github.com/isyll/go-grpc-starter/internal/settings"
	"github.com/isyll/go-grpc-starter/internal/users"
	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
	apptoken "github.com/isyll/go-grpc-starter/pkg/token"
)

// ---- fakes -----------------------------------------------------------------

type fakeUsers struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]*users.User
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byID: map[int64]*users.User{}}
}

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", email)
}

func (f *fakeUsers) ExistsByEmail(ctx context.Context, email string) bool {
	_, err := f.FindByEmail(ctx, email)
	return err == nil
}

func (f *fakeUsers) Create(_ context.Context, user *users.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	user.ID = f.nextID
	cp := *user
	f.byID[user.ID] = &cp
	return nil
}

func (f *fakeUsers) FindByID(_ context.Context, id int64) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, fmt.Errorf("user %d not found", id)
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUsers) UpdateLastLogin(_ context.Context, _ int64) error { return nil }

func (f *fakeUsers) UpdatePasswordHash(_ context.Context, id int64, hash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}
	u.PasswordHash = hash
	return nil
}

func (f *fakeUsers) MarkEmailVerified(_ context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}
	now := time.Now()
	u.EmailVerifiedAt = &now
	return nil
}

type fakeSettings struct{}

func (fakeSettings) Create(context.Context, *settings.UserSettings) error { return nil }
func (fakeSettings) GetByUserID(context.Context, int64) (*settings.Settings, error) {
	s := settings.DefaultSettings()
	return &s, nil
}

type fakeSessions struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]*DeviceSession
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{byID: map[int64]*DeviceSession{}}
}

func (f *fakeSessions) Create(_ context.Context, s *DeviceSession) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	s.ID = f.nextID
	s.LastActivity = time.Now().UTC()
	cp := *s
	f.byID[s.ID] = &cp
	return nil
}

func (f *fakeSessions) FindByID(_ context.Context, id int64) (*DeviceSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeSessions) FindByUserAndDeviceID(
	_ context.Context, userID int64, deviceID string,
) (*DeviceSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.byID {
		if s.UserID == userID && s.DeviceID == deviceID && s.RevokedAt == nil {
			cp := *s
			return &cp, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (f *fakeSessions) Revoke(_ context.Context, reason string, id int64) (*DeviceSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	now := time.Now()
	s.RevokedAt = &now
	s.RevokedReason = reason
	cp := *s
	return &cp, nil
}

func (f *fakeSessions) FindActiveDevicesByUser(
	_ context.Context, userID int64, _ time.Duration,
) ([]DeviceSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []DeviceSession
	for _, s := range f.byID {
		if s.UserID == userID && s.RevokedAt == nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeSessions) RevokeAllByUserID(ctx context.Context, userID int64, reason string) error {
	f.mu.Lock()
	ids := make([]int64, 0)
	for _, s := range f.byID {
		if s.UserID == userID && s.RevokedAt == nil {
			ids = append(ids, s.ID)
		}
	}
	f.mu.Unlock()
	for _, id := range ids {
		_, _ = f.Revoke(ctx, reason, id)
	}
	return nil
}

func (f *fakeSessions) activeCount(userID int64) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, s := range f.byID {
		if s.UserID == userID && s.RevokedAt == nil {
			n++
		}
	}
	return n
}

type fakeRefresh struct {
	mu       sync.Mutex
	byHash   map[string]*RefreshToken
	sessions *fakeSessions
	users    *fakeUsers
}

func newFakeRefresh(sessions *fakeSessions, users *fakeUsers) *fakeRefresh {
	return &fakeRefresh{byHash: map[string]*RefreshToken{}, sessions: sessions, users: users}
}

func (f *fakeRefresh) Create(_ context.Context, t *RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *t
	f.byHash[t.TokenHash] = &cp
	return nil
}

func (f *fakeRefresh) FindByTokenHash(ctx context.Context, hash string) (*RefreshToken, error) {
	f.mu.Lock()
	t, ok := f.byHash[hash]
	if !ok {
		f.mu.Unlock()
		return nil, ErrInvalidToken
	}
	cp := *t
	f.mu.Unlock()

	// Mirror the real repository: the token carries its session and user.
	session, err := f.sessions.FindByID(ctx, cp.SessionID)
	if err != nil {
		return nil, err
	}
	user, err := f.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	session.User = *user
	cp.Session = *session
	return &cp, nil
}

func (f *fakeRefresh) RevokeByTokenHash(_ context.Context, hash, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.byHash[hash]; ok && t.RevokedAt == nil {
		now := time.Now()
		t.RevokedAt = &now
		t.RevokedReason = reason
	}
	return nil
}

func (f *fakeRefresh) RevokeBySessionID(_ context.Context, sessionID int64, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.byHash {
		if t.SessionID == sessionID && t.RevokedAt == nil {
			now := time.Now()
			t.RevokedAt = &now
			t.RevokedReason = reason
		}
	}
	return nil
}

func (f *fakeRefresh) RevokeByTokenFamily(_ context.Context, family, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.byHash {
		if t.TokenFamily == family && t.RevokedAt == nil {
			now := time.Now()
			t.RevokedAt = &now
			t.RevokedReason = reason
		}
	}
	return nil
}

func (f *fakeRefresh) CleanupExpired(context.Context) (int64, error) { return 0, nil }

type fakeEmail struct {
	mu           sync.Mutex
	verification int
	reset        int
}

func (f *fakeEmail) SendVerificationEmail(context.Context, string, string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.verification++
	return nil
}

func (f *fakeEmail) SendPasswordResetEmail(context.Context, string, string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reset++
	return nil
}

type passthroughTx struct{}

func (passthroughTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeTokens struct {
	mu     sync.Mutex
	n      int
	tokens map[string]int64 // token -> session id
}

func newFakeTokens() *fakeTokens { return &fakeTokens{tokens: map[string]int64{}} }

func (f *fakeTokens) Generate(_ context.Context, sessionID, userID int64, deviceID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	tok := fmt.Sprintf("access-%d-%d-%s-%d", sessionID, userID, deviceID, f.n)
	f.tokens[tok] = sessionID
	return tok, nil
}

func (f *fakeTokens) Validate(_ context.Context, token string) (*apptoken.AccessTokenClaims, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sessionID, ok := f.tokens[token]
	if !ok {
		return nil, apptoken.ErrTokenNotFound
	}
	return &apptoken.AccessTokenClaims{SessionID: sessionID}, nil
}

func (f *fakeTokens) Revoke(_ context.Context, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.tokens, token)
	return nil
}

// ---- harness ---------------------------------------------------------------

type harness struct {
	svc      *Service
	users    *fakeUsers
	sessions *fakeSessions
	refresh  *fakeRefresh
	email    *fakeEmail
	cache    *cache.CacheManager
	redis    *miniredis.Miniredis
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cm := cache.NewCacheManager(client, "test")

	cfg := &config.Configs{Security: &config.SecurityConfig{}}
	cfg.Security.Auth.MaxInactivityTimeout = 24 * time.Hour
	cfg.Security.Auth.MaxDevicesPerUser = 5
	cfg.Security.Auth.OAT.AccessTokenExpiry = 30 * time.Minute
	cfg.Security.Auth.OAT.RefreshTokenExpiry = 720 * time.Hour
	cfg.Security.Auth.Lockout.MaxAttempts = 3
	cfg.Security.Auth.Lockout.Window = time.Minute
	// Cheap argon2 parameters keep the suite fast.
	cfg.Security.PasswordHash.Memory = 8 * 1024
	cfg.Security.PasswordHash.Iterations = 1
	cfg.Security.PasswordHash.Parallelism = 1
	cfg.Security.PasswordHash.SaltLength = 16
	cfg.Security.PasswordHash.KeyLength = 32

	logx := logger.New("test")
	fu := newFakeUsers()
	fs := newFakeSessions()
	fr := newFakeRefresh(fs, fu)
	fe := &fakeEmail{}

	svc := NewService(
		cfg, logx,
		newFakeTokens(),
		cm, passthroughTx{}, fu, fs, fakeSettings{}, fr, fe,
		event.New(nil, logx),
	)
	return &harness{svc: svc, users: fu, sessions: fs, refresh: fr, email: fe, cache: cm, redis: mr}
}

// ---- tests -----------------------------------------------------------------

var device = DeviceInfo{DeviceID: "dev-1", Name: "Test Phone", Platform: "android"}

func register(t *testing.T, h *harness, email string) *TokenPair {
	t.Helper()
	tokens, err := h.svc.Register(context.Background(), RegisterInput{
		Email:     email,
		Password:  "correct horse battery",
		FirstName: "Ada",
		LastName:  "Lovelace",
		Device:    device,
	})
	require.NoError(t, err)
	return tokens
}

func TestRegisterIssuesSessionAndTokens(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")

	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	require.NotNil(t, tokens.User)
	assert.Equal(t, "ada@example.com", tokens.User.Email)
	assert.Equal(t, 1, h.sessions.activeCount(tokens.User.ID))
	assert.Equal(t, 1, h.email.verification)
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	h := newHarness(t)
	register(t, h, "ada@example.com")

	_, err := h.svc.Register(context.Background(), RegisterInput{
		Email:    "ada@example.com",
		Password: "another password",
		Device:   device,
	})
	assert.ErrorIs(t, err, ErrEmailExists)
}

func TestLoginSucceedsWithCorrectPassword(t *testing.T) {
	h := newHarness(t)
	register(t, h, "ada@example.com")

	tokens, err := h.svc.Login(context.Background(), LoginInput{
		Email:    "Ada@Example.com ", // normalization: case and whitespace
		Password: "correct horse battery",
		Device:   device,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
}

func TestLoginRejectsWrongPasswordAndUnknownEmail(t *testing.T) {
	h := newHarness(t)
	register(t, h, "ada@example.com")

	_, err := h.svc.Login(context.Background(), LoginInput{
		Email: "ada@example.com", Password: "wrong", Device: device,
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = h.svc.Login(context.Background(), LoginInput{
		Email: "nobody@example.com", Password: "whatever", Device: device,
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLoginLockoutBlocksAfterMaxFailures(t *testing.T) {
	h := newHarness(t)
	register(t, h, "ada@example.com")
	ctx := context.Background()

	for range 3 {
		_, err := h.svc.Login(ctx, LoginInput{
			Email: "ada@example.com", Password: "wrong", Device: device,
		})
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	}

	// Even the correct password is rejected while locked.
	_, err := h.svc.Login(ctx, LoginInput{
		Email: "ada@example.com", Password: "correct horse battery", Device: device,
	})
	assert.ErrorIs(t, err, ErrTooManyAttempts)

	// The window expiring unlocks the account.
	h.redis.FastForward(2 * time.Minute)
	_, err = h.svc.Login(ctx, LoginInput{
		Email: "ada@example.com", Password: "correct horse battery", Device: device,
	})
	assert.NoError(t, err)
}

func TestLoginSuccessResetsLockoutCounter(t *testing.T) {
	h := newHarness(t)
	register(t, h, "ada@example.com")
	ctx := context.Background()

	for range 2 {
		_, _ = h.svc.Login(ctx, LoginInput{
			Email: "ada@example.com", Password: "wrong", Device: device,
		})
	}
	_, err := h.svc.Login(ctx, LoginInput{
		Email: "ada@example.com", Password: "correct horse battery", Device: device,
	})
	require.NoError(t, err)

	// The counter restarted: two more failures stay under the limit.
	for range 2 {
		_, err = h.svc.Login(ctx, LoginInput{
			Email: "ada@example.com", Password: "wrong", Device: device,
		})
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	rotated, err := h.svc.RefreshTokens(ctx, tokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, tokens.RefreshToken, rotated.RefreshToken)
	assert.NotEmpty(t, rotated.AccessToken)
}

func TestRefreshReuseRevokesFamily(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	rotated, err := h.svc.RefreshTokens(ctx, tokens.RefreshToken)
	require.NoError(t, err)

	// Replaying the consumed token is treated as theft: the whole family dies.
	_, err = h.svc.RefreshTokens(ctx, tokens.RefreshToken)
	assert.ErrorIs(t, err, ErrTokenRevoked)

	_, err = h.svc.RefreshTokens(ctx, rotated.RefreshToken)
	assert.ErrorIs(t, err, ErrTokenRevoked)
}

func TestVerifyEmailIsSingleUse(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	require.NoError(t, h.cache.Set(
		ctx, cache.VerificationTokenKey("tok-1"),
		tokenData{UserID: tokens.User.ID}, cache.CacheShort,
	))

	require.NoError(t, h.svc.VerifyEmail(ctx, "tok-1"))
	u, err := h.users.FindByID(ctx, tokens.User.ID)
	require.NoError(t, err)
	assert.True(t, u.IsEmailVerified())

	assert.ErrorIs(t, h.svc.VerifyEmail(ctx, "tok-1"), ErrInvalidVerificationToken)
}

func TestResetPasswordRevokesAllSessionsAndConsumesToken(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	require.NoError(t, h.cache.Set(
		ctx, cache.PasswordResetKey("reset-1"),
		tokenData{UserID: tokens.User.ID}, cache.CacheShort,
	))

	require.NoError(t, h.svc.ResetPassword(ctx, "reset-1", "brand new password"))
	assert.Equal(t, 0, h.sessions.activeCount(tokens.User.ID))
	assert.ErrorIs(t,
		h.svc.ResetPassword(ctx, "reset-1", "brand new password"),
		ErrInvalidResetToken,
	)

	_, err := h.svc.Login(ctx, LoginInput{
		Email: "ada@example.com", Password: "brand new password", Device: device,
	})
	assert.NoError(t, err)
}

func TestChangePasswordRevokesOtherSessions(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	// A second device logs in.
	other, err := h.svc.Login(ctx, LoginInput{
		Email: "ada@example.com", Password: "correct horse battery",
		Device: DeviceInfo{DeviceID: "dev-2", Platform: "ios"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, h.sessions.activeCount(tokens.User.ID))

	currentSession, err := h.sessions.FindByUserAndDeviceID(ctx, tokens.User.ID, "dev-2")
	require.NoError(t, err)

	require.NoError(t, h.svc.ChangePassword(
		ctx, other.User.ID, currentSession.ID,
		"correct horse battery", "a fresh password",
	))

	assert.Equal(t, 1, h.sessions.activeCount(tokens.User.ID))
	remaining, err := h.sessions.FindByUserAndDeviceID(ctx, tokens.User.ID, "dev-2")
	require.NoError(t, err)
	assert.Equal(t, currentSession.ID, remaining.ID)
}

func TestEvictOldestSessionOverDeviceLimit(t *testing.T) {
	h := newHarness(t)
	tokens := register(t, h, "ada@example.com")
	ctx := context.Background()

	for i := 2; i <= 6; i++ {
		_, err := h.svc.Login(ctx, LoginInput{
			Email: "ada@example.com", Password: "correct horse battery",
			Device: DeviceInfo{DeviceID: fmt.Sprintf("dev-%d", i), Platform: "android"},
		})
		require.NoError(t, err)
	}
	assert.LessOrEqual(t, h.sessions.activeCount(tokens.User.ID), 5)
}
