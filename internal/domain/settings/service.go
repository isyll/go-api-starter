package settings

import (
	"context"

	"github.com/isyll/go-api-starter/internal/models"
)

// Service holds user-settings logic.
type Service struct {
	repo Repository
}

// NewService builds the settings service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Get returns the settings for a user.
func (s *Service) Get(ctx context.Context, userID int64) (*models.Settings, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// Update replaces the settings for a user.
func (s *Service) Update(ctx context.Context, userID int64, settings models.Settings) error {
	return s.repo.Update(ctx, userID, settings)
}
