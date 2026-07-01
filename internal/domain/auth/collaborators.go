package auth

import (
	"context"

	"github.com/isyll/go-api-starter/internal/models"
)

// UserStore is the user data access the auth domain needs.
type UserStore interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) bool
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id int64) (*models.User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
}

// SettingsStore is the user-settings data access the auth domain needs.
type SettingsStore interface {
	Create(ctx context.Context, settings *models.UserSettings) error
	GetByUserID(ctx context.Context, userID int64) (*models.Settings, error)
}

// EmailSender enqueues transactional auth emails. Implemented by an
// adapter over the email worker dispatcher in the app wiring.
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}
