package notifications

import (
	"strings"
	"time"
)

// Priority maps to the Asynq queue weights for notification delivery.
// Higher priority tasks are processed first.
type Priority string // @name Priority

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Event is the payload sent to the notification worker. The Type
// field drives template selection and preference-category matching;
// Data holds the dynamic template substitutions.
type Event struct {
	Type           string            `json:"type"`
	UserID         int64             `json:"user_id"`
	Data           map[string]string `json:"data"`
	Priority       Priority          `json:"priority,omitempty"`
	ScheduledAt    *time.Time        `json:"scheduled_at,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
} // @name Event

// SendResult reports whether a single FCM push delivery succeeded.
// On failure, ErrorCode and ErrorMessage carry the provider response.
type SendResult struct {
	Success      bool   `json:"success"`
	FCMTokenID   int64  `json:"fcm_token_id,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
} // @name SendResult

// EventCategory maps event types to their preference category
func EventCategory(eventType string) string {
	switch {
	case strings.HasPrefix(eventType, "booking."):
		return "booking_updates"
	case strings.HasPrefix(eventType, "trip.reminder"):
		return "trip_reminders"
	case strings.HasPrefix(eventType, "trip."):
		return "trip_updates"
	case strings.HasPrefix(eventType, "message."):
		return "messages"
	case strings.HasPrefix(eventType, "rating."):
		return "ratings"
	case strings.HasPrefix(eventType, "marketing."):
		return "marketing"
	default:
		return ""
	}
}
