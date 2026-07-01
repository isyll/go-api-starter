// Package auth handles email/password authentication: registration,
// login, opaque access tokens, refresh-token rotation, device
// sessions, email verification, and password reset.
package auth

import (
	"time"

	"github.com/isyll/go-api-starter/internal/models"
)

// DeviceInfo is the client-supplied device metadata attached to a
// session at login. All fields are optional except when the caller
// wants stable per-device sessions (then DeviceID must be set).
type DeviceInfo struct {
	DeviceID     string
	Name         string
	Platform     string
	Model        string
	Manufacturer string
	IPAddress    string
	UserAgent    string
}

func (d DeviceInfo) toSession(userID int64) *models.DeviceSession {
	return &models.DeviceSession{
		UserID:       userID,
		DeviceID:     d.DeviceID,
		Name:         d.Name,
		Platform:     d.Platform,
		Model:        d.Model,
		Manufacturer: d.Manufacturer,
		IPAddress:    d.IPAddress,
		UserAgent:    d.UserAgent,
	}
}

// RegisterInput carries the fields needed to create an account.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Device    DeviceInfo
}

// LoginInput carries login credentials plus device metadata.
type LoginInput struct {
	Email    string
	Password string
	Device   DeviceInfo
}

// TokenPair is the result of a successful authentication.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *models.User
	Settings     *models.Settings
}

// DeviceSessionInfo is a summary of one active session, used by the
// list-devices endpoint.
type DeviceSessionInfo struct {
	ID           int64
	DeviceID     string
	DeviceName   string
	Platform     string
	Manufacturer string
	Model        string
	LastUsedAt   time.Time
	Current      bool
}
