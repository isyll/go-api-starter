package settings

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID int64) (*Settings, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID int64, settings Settings) error {
	return s.repo.Update(ctx, userID, settings)
}
