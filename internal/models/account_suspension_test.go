package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/isyll/go-grpc-starter/internal/models"
)

func TestAccountSuspension_IsActive(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	future := time.Now().Add(2 * time.Hour)

	tests := []struct {
		name     string
		susp     models.AccountSuspension
		expected bool
	}{
		{
			name: "permanent suspension is always active",
			susp: models.AccountSuspension{
				IsPermanent: true,
			},
			expected: true,
		},
		{
			name: "temporary suspension active if until is in future",
			susp: models.AccountSuspension{
				IsPermanent:    false,
				SuspendedUntil: &future,
			},
			expected: true,
		},
		{
			name: "temporary suspension inactive if until is in past",
			susp: models.AccountSuspension{
				IsPermanent:    false,
				SuspendedUntil: &past,
			},
			expected: false,
		},
		{
			name: "temporary suspension inactive if until is null",
			susp: models.AccountSuspension{
				IsPermanent:    false,
				SuspendedUntil: nil,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.susp.IsActive())
		})
	}
}

func TestAccountSuspension_BeforeCreate(t *testing.T) {
	susp := models.AccountSuspension{}

	err := susp.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)

	assert.False(t, susp.SuspendedAt.IsZero())

	// Test if it doesn't override existing value
	specificTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	susp2 := models.AccountSuspension{
		SuspendedAt: specificTime,
	}

	err = susp2.BeforeCreate(&gorm.DB{})
	require.NoError(t, err)
	assert.Equal(t, specificTime, susp2.SuspendedAt)
}
