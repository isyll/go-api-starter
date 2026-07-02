package utils

import "time"

func IntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func Int64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func IntValue(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func TimeValue(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

func BoolValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func IntValueOr(p *int, defaultValue int) int {
	if p == nil {
		return defaultValue
	}
	return *p
}

func StringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
