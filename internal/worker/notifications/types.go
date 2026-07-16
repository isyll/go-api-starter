package notifications

import (
	"strings"
	"time"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// queueName namespaces queues so only this worker consumes them.
func queueName(p Priority) string {
	return "notifications:" + string(p)
}

func QueueNames() []string {
	return []string{
		queueName(PriorityHigh),
		queueName(PriorityNormal),
		queueName(PriorityLow),
	}
}

type Event struct {
	Type           string            `json:"type"`
	UserID         int64             `json:"user_id"`
	Data           map[string]string `json:"data"`
	Priority       Priority          `json:"priority,omitempty"`
	ScheduledAt    *time.Time        `json:"scheduled_at,omitempty"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
}

type SendResult struct {
	Success      bool   `json:"success"`
	FCMTokenID   int64  `json:"fcm_token_id,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func EventCategory(eventType string) string {
	if strings.HasPrefix(eventType, "marketing.") {
		return "marketing"
	}
	return ""
}
