package suspension

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
	"github.com/isyll/go-api-starter/internal/models"
)

// ErrUntilRequired - 400. A non-permanent suspension needs an expiry.
var ErrUntilRequired = apperrors.BadRequest(codes.InvalidParam, "suspension.until_required")

// Service exposes suspension lookups and admin suspend/unsuspend.
type Service struct {
	repo Repository
}

// NewService builds the suspension service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetActive returns the active suspension for a user, or nil when the
// account is in good standing.
func (s *Service) GetActive(ctx context.Context, userID int64) (*models.AccountSuspension, error) {
	sus, err := s.repo.GetActiveByUserID(ctx, userID)
	if errors.Is(err, ErrNotSuspended) {
		return nil, nil
	}
	return sus, err
}

// SuspendInput describes a new suspension.
type SuspendInput struct {
	UserID    int64
	Reason    models.SuspensionReason
	Details   string
	Until     *time.Time
	Permanent bool
}

// Suspend records a suspension. Either Permanent or Until must be set.
func (s *Service) Suspend(ctx context.Context, in SuspendInput) (*models.AccountSuspension, error) {
	if !in.Permanent && in.Until == nil {
		return nil, ErrUntilRequired
	}
	row := &models.AccountSuspension{
		UserID:         in.UserID,
		Reason:         in.Reason,
		Details:        in.Details,
		SuspendedUntil: in.Until,
		IsPermanent:    in.Permanent,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

// Unsuspend ends every active suspension for a user.
func (s *Service) Unsuspend(ctx context.Context, userID int64) error {
	return s.repo.DeactivateByUserID(ctx, userID)
}
