package users

import (
	"context"
	"time"

	"github.com/isyll/go-api-starter/internal/events"
	"github.com/isyll/go-api-starter/internal/models"
	"github.com/isyll/go-api-starter/pkg/logger"
)

// ProfileUpdate carries optional profile fields. Nil fields are left
// unchanged.
type ProfileUpdate struct {
	FirstName *string
	LastName  *string
	Bio       *string
	Avatar    *string
}

// Service holds user profile and account-lifecycle logic.
type Service struct {
	repo     Repository
	sessions SessionRevoker
	bus      *events.Bus
	logger   *logger.Logger
}

// NewService builds the users service.
func NewService(
	repo Repository,
	sessions SessionRevoker,
	bus *events.Bus,
	logx *logger.Logger,
) *Service {
	return &Service{repo: repo, sessions: sessions, bus: bus, logger: logx}
}

// Get returns a user by id.
func (s *Service) Get(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateProfile applies the non-nil fields of upd to the user.
func (s *Service) UpdateProfile(
	ctx context.Context, id int64, upd ProfileUpdate,
) (*models.User, error) {
	fields := map[string]any{}
	if upd.FirstName != nil {
		fields["first_name"] = *upd.FirstName
	}
	if upd.LastName != nil {
		fields["last_name"] = *upd.LastName
	}
	if upd.Bio != nil {
		fields["bio"] = *upd.Bio
	}
	if upd.Avatar != nil {
		fields["avatar"] = *upd.Avatar
	}
	return s.repo.UpdateProfile(ctx, id, fields)
}

// DeleteAccount soft-deletes the user, revokes all sessions, and
// publishes UserAccountDeleted so caches drop the account.
func (s *Service) DeleteAccount(ctx context.Context, id int64) error {
	if err := s.repo.SoftDeleteByID(ctx, id); err != nil {
		return err
	}
	if err := s.sessions.RevokeAllSessions(ctx, id, "account_deleted"); err != nil {
		s.logger.Warn("revoke sessions on delete failed", "error", err, "user_id", id)
	}
	if err := s.bus.Publish(ctx, &events.UserAccountDeleted{
		UserID:     id,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Warn("publish account deleted failed", "error", err, "user_id", id)
	}
	return nil
}

// List returns a page of users and the total count (admin).
func (s *Service) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

// SetStatus changes a user's account status (admin).
func (s *Service) SetStatus(ctx context.Context, id int64, status models.UserStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// SetRole changes a user's role (admin).
func (s *Service) SetRole(ctx context.Context, id int64, role models.UserRole) error {
	return s.repo.UpdateRole(ctx, id, role)
}
