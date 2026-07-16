// Package webhooks carries inbound provider webhooks to a background worker.
package webhooks

import "time"

const TaskWebhookReceived = "webhook:received"

const queueWebhooks = "webhooks:default"

func QueueNames() []string {
	return []string{queueWebhooks}
}

// ReceivedEvent is a verified webhook; Payload is the raw provider body.
type ReceivedEvent struct {
	Provider   string            `json:"provider"`
	Payload    []byte            `json:"payload"`
	Headers    map[string]string `json:"headers,omitempty"`
	ReceivedAt time.Time         `json:"received_at"`
	RequestID  string            `json:"request_id,omitempty"`
}
