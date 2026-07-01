package models

import (
	"time"

	"github.com/isyll/go-api-starter/pkg/utils"

	"gorm.io/gorm"
)

// RefreshToken is the GORM model for auth.refresh_tokens. The raw
// token string is never stored; only its SHA-256 hash is persisted in
// TokenHash. Token family tracking detects replay attacks — presenting
// a revoked token from the same family invalidates the entire family.
type RefreshToken struct {
	ID        string `gorm:"primaryKey" json:"id"                       msgpack:"id"`
	SessionID int64  `                  json:"session_id"               msgpack:"session_id"`
	// TokenHash is the SHA-256 hex digest of the raw refresh token.
	// The raw token is never persisted.
	TokenHash string `                  json:"token_hash"               msgpack:"token_hash"`
	// TokenPrefix holds the first 8 hex characters of TokenHash.
	// It is indexed for cheap prefix-based lookup so the application can
	// narrow the row set before performing a constant-time full-hash
	// comparison, avoiding both timing leaks and full-table scans.
	TokenPrefix string `gorm:"column:token_prefix" json:"token_prefix" msgpack:"token_prefix"`
	// TokenFamily groups tokens issued in the same rotation chain.
	// When a revoked token is presented, all tokens in the family are
	// revoked to neutralize replay attacks.
	TokenFamily   string     `                  json:"token_family"             msgpack:"token_family"`
	ExpiresAt     time.Time  `                  json:"expires_at"               msgpack:"expires_at"`
	RevokedAt     *time.Time `                  json:"revoked_at,omitempty"     msgpack:"revoked_at,omitempty"`
	RevokedReason string     `                  json:"revoked_reason,omitempty" msgpack:"revoked_reason,omitempty"`
	CreatedAt     time.Time  `                  json:"created_at"               msgpack:"created_at"`

	Session DeviceSession `gorm:"foreignKey:SessionID" json:"session,omitempty" msgpack:"session,omitempty"`
}

// BeforeCreate generates a UUID for the token ID when not set and
// populates TokenPrefix from the first 8 hex characters of TokenHash.
// TokenPrefix is used for indexed prefix lookup during token refresh.
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == "" {
		rt.ID = utils.NewUUIDNoDash()
	}
	if rt.TokenPrefix == "" && len(rt.TokenHash) >= 8 {
		rt.TokenPrefix = rt.TokenHash[:8]
	}
	return nil
}

// TableName returns the schema-qualified table name for GORM.
func (RefreshToken) TableName() string {
	return "auth.refresh_tokens"
}

// IsValid reports whether the token is neither revoked nor expired.
func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil &&
		rt.ExpiresAt.After(time.Now())
}

// IsRevoked reports whether the token was explicitly revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil && !rt.RevokedAt.IsZero()
}

// IsExpired reports whether the token has passed its ExpiresAt
// timestamp.
func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.Before(time.Now())
}
