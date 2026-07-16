package users

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Avatar       string
	Bio          string
	Status       UserStatus
	Role         UserRole

	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func (u *User) IsActive() bool        { return u.Status == UserStatusActive }
func (u *User) IsSuspended() bool     { return u.Status == UserStatusSuspended }
func (u *User) IsAdmin() bool         { return u.Role == UserRoleAdmin }
func (u *User) IsEmailVerified() bool { return u.EmailVerifiedAt != nil }
