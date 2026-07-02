package auth

import (
	"context"

	"github.com/isyll/go-grpc-starter/internal/models"
)

type UserStore interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) bool
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int64) (*models.User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
}

type SettingsStore interface {
	Create(ctx context.Context, settings *models.UserSettings) error
	GetByUserID(ctx context.Context, userID int64) (*models.Settings, error)
}

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}
