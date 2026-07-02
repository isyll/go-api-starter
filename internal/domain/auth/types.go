// Package auth handles email/password authentication and sessions.
package auth

import (
	"time"

	"github.com/isyll/go-grpc-starter/internal/models"
)

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

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Device    DeviceInfo
}

type LoginInput struct {
	Email    string
	Password string
	Device   DeviceInfo
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *models.User
	Settings     *models.Settings
}

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
