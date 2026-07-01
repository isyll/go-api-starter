package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-api-starter/internal/models"
)

func TestDeviceSession_IsRevoked(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
		expected  bool
	}{
		{
			name:      "not revoked when nil",
			revokedAt: nil,
			expected:  false,
		},
		{
			name:      "revoked when set",
			revokedAt: ptrTime(time.Now()),
			expected:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ds := &models.DeviceSession{RevokedAt: tc.revokedAt}
			assert.Equal(t, tc.expected, ds.IsRevoked())
		})
	}
}

func TestDeviceSession_IsInactive(t *testing.T) {
	timeout := 30 * time.Minute

	t.Run("active when recent", func(t *testing.T) {
		ds := &models.DeviceSession{
			LastActivity: time.Now().Add(-10 * time.Minute),
		}
		assert.False(t, ds.IsInactive(timeout))
	})

	t.Run("inactive when old", func(t *testing.T) {
		ds := &models.DeviceSession{
			LastActivity: time.Now().Add(-60 * time.Minute),
		}
		assert.True(t, ds.IsInactive(timeout))
	})
}

func TestDeviceSession_IsValid(t *testing.T) {
	timeout := 30 * time.Minute

	t.Run("valid when active and not revoked", func(t *testing.T) {
		ds := &models.DeviceSession{
			LastActivity: time.Now(),
		}
		assert.True(t, ds.IsValid(timeout))
	})

	t.Run("invalid when revoked", func(t *testing.T) {
		now := time.Now()
		ds := &models.DeviceSession{
			LastActivity: now,
			RevokedAt:    &now,
		}
		assert.False(t, ds.IsValid(timeout))
	})

	t.Run("invalid when inactive", func(t *testing.T) {
		ds := &models.DeviceSession{
			LastActivity: time.Now().Add(-60 * time.Minute),
		}
		assert.False(t, ds.IsValid(timeout))
	})
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
