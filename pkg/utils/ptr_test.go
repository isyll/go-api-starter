package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-api-starter/pkg/utils"
)

func TestIntPtr(t *testing.T) {
	t.Run("returns nil for zero", func(t *testing.T) {
		result := utils.IntPtr(0)
		assert.Nil(t, result)
	})

	t.Run("returns pointer for non-zero", func(t *testing.T) {
		result := utils.IntPtr(42)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("returns pointer for negative", func(t *testing.T) {
		result := utils.IntPtr(-10)
		assert.NotNil(t, result)
		assert.Equal(t, -10, *result)
	})
}

func TestUintPtr(t *testing.T) {
	t.Run("returns nil for zero", func(t *testing.T) {
		result := utils.Int64Ptr(0)
		assert.Nil(t, result)
	})

	t.Run("returns pointer for non-zero", func(t *testing.T) {
		result := utils.Int64Ptr(100)
		assert.NotNil(t, result)
		assert.Equal(t, int64(100), *result)
	})
}

func TestIntValue(t *testing.T) {
	t.Run("returns zero for nil", func(t *testing.T) {
		result := utils.IntValue(nil)
		assert.Equal(t, 0, result)
	})

	t.Run("returns value for non-nil", func(t *testing.T) {
		v := 42
		result := utils.IntValue(&v)
		assert.Equal(t, 42, result)
	})
}

func TestTimeValue(t *testing.T) {
	t.Run("returns zero time for nil", func(t *testing.T) {
		result := utils.TimeValue(nil)
		assert.True(t, result.IsZero())
	})

	t.Run("returns value for non-nil", func(t *testing.T) {
		now := time.Now()
		result := utils.TimeValue(&now)
		assert.Equal(t, now, result)
	})
}

func TestBoolValue(t *testing.T) {
	t.Run("returns false for nil", func(t *testing.T) {
		result := utils.BoolValue(nil)
		assert.False(t, result)
	})

	t.Run("returns true for true pointer", func(t *testing.T) {
		v := true
		result := utils.BoolValue(&v)
		assert.True(t, result)
	})

	t.Run("returns false for false pointer", func(t *testing.T) {
		v := false
		result := utils.BoolValue(&v)
		assert.False(t, result)
	})
}

func TestIntValueOr(t *testing.T) {
	t.Run("returns default for nil", func(t *testing.T) {
		result := utils.IntValueOr(nil, 99)
		assert.Equal(t, 99, result)
	})

	t.Run("returns value when not nil", func(t *testing.T) {
		v := 42
		result := utils.IntValueOr(&v, 99)
		assert.Equal(t, 42, result)
	})
}

func TestStringValue(t *testing.T) {
	t.Run("returns empty for nil", func(t *testing.T) {
		result := utils.StringValue(nil)
		assert.Empty(t, result)
	})

	t.Run("returns value for non-nil", func(t *testing.T) {
		v := "hello"
		result := utils.StringValue(&v)
		assert.Equal(t, "hello", result)
	})
}
