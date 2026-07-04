package suspension

import (
	"context"
	"errors"
	"time"

	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var ErrUntilRequired = errs.BadRequest(codes.InvalidParam, "suspension.until_required")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetActive(ctx context.Context, userID int64) (*AccountSuspension, error) {
	sus, err := s.repo.GetActiveByUserID(ctx, userID)
	if errors.Is(err, ErrNotSuspended) {
		return nil, nil
	}
	return sus, err
}

type SuspendInput struct {
	UserID    int64
	Reason    SuspensionReason
	Details   string
	Until     *time.Time
	Permanent bool
}

func (s *Service) Suspend(ctx context.Context, in SuspendInput) (*AccountSuspension, error) {
	if !in.Permanent && in.Until == nil {
		return nil, ErrUntilRequired
	}
	row := &AccountSuspension{
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
