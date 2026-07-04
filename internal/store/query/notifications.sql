-- name: GetNotificationTemplateByEventType :one
SELECT * FROM notifications.notification_templates WHERE event_type = $1;

-- name: ListTemplateTranslations :many
SELECT * FROM notifications.notification_template_translations
WHERE template_id = $1;

-- name: CreateNotificationLog :one
INSERT INTO notifications.notification_logs (
  user_id, event_type, event_id, fcm_token_id,
  status, error_code, error_message, payload
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
