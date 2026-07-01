package models

import (
	"fmt"
	"time"

	"github.com/isyll/go-api-starter/pkg/idenc"

	"gorm.io/gorm"
)

// UserStatus is the account lifecycle state. The auth middleware
// checks this value and blocks non-active accounts.
type UserStatus string // @name UserStatus

const (
	// UserStatusActive means the account is in good standing.
	UserStatusActive UserStatus = "active"
	// UserStatusInactive means the user self-deactivated. It can
	// be reactivated later.
	UserStatusInactive UserStatus = "inactive"
	// UserStatusSuspended means an admin suspended the account.
	UserStatusSuspended UserStatus = "suspended"
)

// UserRole controls access to admin-only features.
type UserRole string // @name UserRole

const (
	// UserRoleUser is a normal end user.
	UserRoleUser UserRole = "user"
	// UserRoleAdmin can access admin-only endpoints.
	UserRoleAdmin UserRole = "admin"
)

// User is the GORM model for auth.users. Email and password are the
// login credentials. Row-level security limits reads and writes to
// the owner row plus admin roles.
type User struct {
	ID    int64  `gorm:"primaryKey" json:"id"    msgpack:"id"`
	Email string `                  json:"email" msgpack:"email"`
	// PasswordHash is the bcrypt hash. Never serialized.
	PasswordHash string `json:"-" msgpack:"-"`

	FirstName string `json:"first_name" msgpack:"first_name"`
	LastName  string `json:"last_name"  msgpack:"last_name"`
	Avatar    string `json:"avatar"     msgpack:"avatar"`
	Bio       string `json:"bio"        msgpack:"bio"`

	// Status is the account lifecycle state. Defaults to active.
	Status UserStatus `json:"status" msgpack:"status"`
	// Role controls admin access. Defaults to user.
	Role UserRole `json:"role" msgpack:"role"`

	// EmailVerifiedAt is set when the user confirms their email.
	EmailVerifiedAt *time.Time `json:"email_verified_at" msgpack:"email_verified_at"`
	LastLoginAt     *time.Time `json:"last_login_at"     msgpack:"last_login_at"`

	CreatedAt time.Time `json:"created_at" msgpack:"created_at"`
	UpdatedAt time.Time `json:"updated_at" msgpack:"updated_at"`

	// DeletedAt supports soft-delete. Foreign keys use ON DELETE
	// RESTRICT so a soft-deleted row is never orphaned.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" msgpack:"-"`

	ActiveSuspension *AccountSuspension  `gorm:"foreignKey:UserID;references:ID" json:"active_suspension,omitempty" msgpack:"active_suspension,omitempty"`
	StatusHistory    []UserStatusHistory `gorm:"foreignKey:UserID"               json:"status_history,omitempty"    msgpack:"status_history,omitempty"`
	UserSettings     *UserSettings       `gorm:"foreignKey:UserID;references:ID" json:"user_settings,omitempty"     msgpack:"user_settings,omitempty"`
}

type UserList []*User

func (User) TableName() string {
	return "auth.users"
}

// BeforeCreate applies default status and role when absent.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	if u.Role == "" {
		u.Role = UserRoleUser
	}
	return nil
}

// IsActive reports whether the account status is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsSuspended reports whether the account status is suspended.
func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

// IsAdmin reports whether the user has the admin role.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsEmailVerified reports whether the user confirmed their email.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// HasActiveSuspension reports whether the user is suspended and has a
// currently running suspension record loaded.
func (u *User) HasActiveSuspension() bool {
	if !u.IsSuspended() || u.ActiveSuspension == nil {
		return false
	}
	return u.ActiveSuspension.IsActive()
}

// GetFullName returns "FirstName LastName".
func (u *User) GetFullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// ToResponse builds the owner view of a user, encoding the internal
// ID via the encoder.
func (u *User) ToResponse(encoder idenc.IDEncoder) *UserResponse {
	return &UserResponse{
		ID:            encoder.Encode(u.ID),
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Avatar:        u.Avatar,
		Bio:           u.Bio,
		Status:        u.Status,
		Role:          u.Role,
		EmailVerified: u.IsEmailVerified(),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// ToPublicProfile builds the public view, omitting email and status.
func (u *User) ToPublicProfile(encoder idenc.IDEncoder) *UserPublicProfile {
	return &UserPublicProfile{
		ID:        encoder.Encode(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		CreatedAt: u.CreatedAt,
	}
}

// UserResponse is the full owner view returned from owner-scoped
// endpoints.
type UserResponse struct {
	ID            string     `json:"id"             msgpack:"id"             example:"18n7q8765"`
	Email         string     `json:"email"          msgpack:"email"          example:"john@example.com"`
	FirstName     string     `json:"first_name"     msgpack:"first_name"     example:"John"`
	LastName      string     `json:"last_name"      msgpack:"last_name"      example:"Doe"`
	Avatar        string     `json:"avatar"         msgpack:"avatar"`
	Bio           string     `json:"bio"            msgpack:"bio"`
	Status        UserStatus `json:"status"         msgpack:"status"         example:"active"`
	Role          UserRole   `json:"role"           msgpack:"role"           example:"user"`
	EmailVerified bool       `json:"email_verified" msgpack:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"     msgpack:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"     msgpack:"updated_at"`
} // @name UserResponse

// UserPublicProfile is the publicly shareable subset of a profile.
type UserPublicProfile struct {
	ID        string    `json:"id"           msgpack:"id"           example:"18n7q8765"`
	FirstName string    `json:"first_name"   msgpack:"first_name"   example:"John"`
	LastName  string    `json:"last_name"    msgpack:"last_name"    example:"Doe"`
	Avatar    string    `json:"avatar"       msgpack:"avatar"`
	Bio       string    `json:"bio"          msgpack:"bio"`
	CreatedAt time.Time `json:"member_since" msgpack:"member_since"`
} // @name UserPublicProfile
