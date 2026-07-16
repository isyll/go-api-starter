package auth

import (
	"time"

	"github.com/isyll/go-grpc-starter/internal/users"
)

type DeviceSession struct {
	ID               int64
	UserID           int64
	Platform         string
	Manufacturer     string
	Model            string
	Version          string
	SDK              string
	Brand            string
	Hardware         string
	Board            string
	Device           string
	Product          string
	IsPhysicalDevice bool
	Name             string
	Identifier       string
	DeviceID         string
	LastActivity     time.Time
	IPAddress        string
	UserAgent        string
	Location         string
	RevokedAt        *time.Time
	RevokedReason    string
	CreatedAt        time.Time

	User users.User
}

func (ds *DeviceSession) IsRevoked() bool {
	return ds.RevokedAt != nil && !ds.RevokedAt.IsZero()
}

func (ds *DeviceSession) IsInactive(timeout time.Duration) bool {
	return time.Since(ds.LastActivity) > timeout
}

func (ds *DeviceSession) IsValid(timeout time.Duration) bool {
	return !ds.IsRevoked() && !ds.IsInactive(timeout)
}

type RefreshToken struct {
	ID            string
	SessionID     int64
	TokenHash     string
	TokenPrefix   string
	TokenFamily   string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	RevokedReason string
	CreatedAt     time.Time

	Session DeviceSession
}

func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil && !rt.RevokedAt.IsZero()
}

func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.Before(time.Now())
}
