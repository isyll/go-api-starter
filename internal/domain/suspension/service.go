package suspension

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/isyll/go-grpc-starter/internal/errors"
	"github.com/isyll/go-grpc-starter/internal/errors/codes"
	"github.com/isyll/go-grpc-starter/internal/models"
)

var ErrUntilRequired = apperrors.BadRequest(codes.InvalidParam, "suspension.until_required")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetActive(ctx context.Context, userID int64) (*models.AccountSuspension, error) {
	sus, err := s.repo.GetActiveByUserID(ctx, userID)
	if errors.Is(err, ErrNotSuspended) {
		return nil, nil
	}
	return sus, err
}

type SuspendInput struct {
	UserID    int64
	Reason    models.SuspensionReason
	Details   string
	Until     *time.Time
	Permanent bool
}

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

func (s *Service) Unsuspend(ctx context.Context, userID int64) error {
	return s.repo.DeactivateByUserID(ctx, userID)
}
