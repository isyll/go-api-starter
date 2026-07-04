package auth

import (
	"context"

	"github.com/isyll/go-grpc-starter/internal/settings"
	"github.com/isyll/go-grpc-starter/internal/users"
)

type UserStore interface {
	FindByEmail(ctx context.Context, email string) (*users.User, error)
	ExistsByEmail(ctx context.Context, email string) bool
	Create(ctx context.Context, user *users.User) error
	FindByID(ctx context.Context, id int64) (*users.User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
	MarkEmailVerified(ctx context.Context, id int64) error
}

type SettingsStore interface {
	Create(ctx context.Context, settings *settings.UserSettings) error
	GetByUserID(ctx context.Context, userID int64) (*settings.Settings, error)
}

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}
