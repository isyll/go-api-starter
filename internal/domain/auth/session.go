package auth

import (
	"context"
	"fmt"

	"github.com/isyll/go-api-starter/internal/infra/cache"
	"github.com/isyll/go-api-starter/internal/models"
)

// createSessionAndTokens opens a new device session and issues a token
// pair. If the user is at the device limit, the oldest session is
// evicted first.
func (s *Service) createSessionAndTokens(
	ctx context.Context,
	user *models.User,
	device DeviceInfo,
	settings *models.Settings,
) (*TokenPair, error) {
	s.evictOldestIfOverLimit(ctx, user.ID)

	session := device.toSession(user.ID)
	s.sessions.Create(ctx, session)
	session.User = *user

	tokens, err := s.generateTokenPair(ctx, user, session, settings)
	if err != nil {
		return nil, err
	}
	if err := s.cacheManager.Set(
		ctx, cache.SessionDataKey(session.ID), session, cache.CacheLong,
	); err != nil {
		s.logger.Warn("session cache write failed", "error", err, "session_id", session.ID)
	}
	return tokens, nil
}

// Logout revokes the session, its access token, and its refresh
// tokens. Errors are logged, not returned.
func (s *Service) Logout(ctx context.Context, sessionID int64, accessToken string) {
	if _, err := s.sessions.Revoke(ctx, "logout", sessionID); err != nil {
		s.logger.Warn("logout: session already revoked", "session_id", sessionID)
	}
	if accessToken != "" {
		if err := s.atManager.Revoke(ctx, accessToken); err != nil {
			s.logger.Error("logout: revoke access token failed", "error", err)
		}
	}
	if err := s.refresh.RevokeBySessionID(ctx, sessionID, "logout"); err != nil {
		s.logger.Error("logout: revoke refresh tokens failed", "error", err)
	}
	_ = s.cacheManager.Delete(ctx, cache.SessionDataKey(sessionID))
}

// ListDevices returns the active sessions for a user. The session
// matching currentSessionID is flagged Current.
func (s *Service) ListDevices(
	ctx context.Context,
	userID int64,
	currentSessionID int64,
) []DeviceSessionInfo {
	sessions := s.sessions.FindActiveDevicesByUser(
		ctx, userID, s.cfg.Security.Auth.MaxInactivityTimeout,
	)
	infos := make([]DeviceSessionInfo, len(sessions))
	for i := range sessions {
		infos[i] = DeviceSessionInfo{
			ID:           sessions[i].ID,
			DeviceID:     sessions[i].DeviceID,
			DeviceName:   sessions[i].Name,
			Platform:     sessions[i].Platform,
			Manufacturer: sessions[i].Manufacturer,
			Model:        sessions[i].Model,
			LastUsedAt:   sessions[i].LastActivity,
			Current:      sessions[i].ID == currentSessionID,
		}
	}
	return infos
}

// RemoveDevice revokes a specific device session. It refuses to remove
// the caller's current session (use Logout for that).
func (s *Service) RemoveDevice(
	ctx context.Context,
	userID int64,
	deviceID string,
	currentSessionID int64,
) error {
	session := s.sessions.FindByUserAndDeviceID(ctx, userID, deviceID)
	if session == nil {
		return ErrSessionNotFound
	}
	if session.ID == currentSessionID {
		return ErrCannotRemoveCurrentDevice
	}
	return s.revokeSessionAndTokens(ctx, session, "user_revoked_device")
}

// RevokeAllSessions revokes every active session for a user. Used by
// account deletion and password reset.
func (s *Service) RevokeAllSessions(ctx context.Context, userID int64, reason string) error {
	sessions := s.sessions.FindActiveDevicesByUser(
		ctx, userID, s.cfg.Security.Auth.MaxInactivityTimeout,
	)
	for i := range sessions {
		if err := s.revokeSessionAndTokens(ctx, &sessions[i], reason); err != nil {
			s.logger.Error("revoke session failed", "session_id", sessions[i].ID, "error", err)
		}
	}
	return nil
}

// GetDeviceSession returns the active session for a user+device, or nil.
func (s *Service) GetDeviceSession(
	ctx context.Context,
	userID int64,
	deviceID string,
) *models.DeviceSession {
	return s.sessions.FindByUserAndDeviceID(ctx, userID, deviceID)
}

// evictOldestIfOverLimit revokes the least recently used session when
// the user is already at MaxDevicesPerUser.
func (s *Service) evictOldestIfOverLimit(ctx context.Context, userID int64) {
	max := s.cfg.Security.Auth.MaxDevicesPerUser
	if max <= 0 {
		return
	}
	sessions := s.sessions.FindActiveDevicesByUser(
		ctx, userID, s.cfg.Security.Auth.MaxInactivityTimeout,
	)
	for len(sessions) >= max {
		oldest := &sessions[len(sessions)-1]
		if err := s.revokeSessionAndTokens(ctx, oldest, "device_limit"); err != nil {
			s.logger.Error("evict oldest session failed", "error", err)
			return
		}
		sessions = sessions[:len(sessions)-1]
	}
}

// revokeSessionAndTokens revokes a session, its refresh tokens, and its
// cache entry.
func (s *Service) revokeSessionAndTokens(
	ctx context.Context,
	session *models.DeviceSession,
	reason string,
) error {
	if _, err := s.sessions.Revoke(ctx, reason, session.ID); err != nil {
		return fmt.Errorf("revoke session %d: %w", session.ID, err)
	}
	if err := s.refresh.RevokeBySessionID(ctx, session.ID, reason); err != nil {
		s.logger.Error("revoke refresh tokens failed", "session_id", session.ID, "error", err)
	}
	_ = s.cacheManager.Delete(ctx, cache.SessionDataKey(session.ID))
	return nil
}
