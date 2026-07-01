package utils

import "time"

// IntPtr returns a pointer to i, or nil when i is zero.
func IntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// Int64Ptr returns a pointer to i, or nil when i is zero.
func Int64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

// IntValue dereferences p, returning 0 when p is nil.
func IntValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// TimeValue dereferences p, returning the zero time.Time when p is nil.
func TimeValue(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

// BoolValue dereferences p, returning false when p is nil.
func BoolValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// IntValueOr dereferences p, returning defaultValue when p is nil.
func IntValueOr(p *int, defaultValue int) int {
	if p == nil {
		return defaultValue
	}
	return *p
}

// StringValue dereferences p, returning an empty string when p is nil.
func StringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
