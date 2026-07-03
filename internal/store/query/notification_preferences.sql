-- name: GetNotificationPreferences :one
SELECT * FROM notifications.notification_preferences WHERE user_id = $1;

-- name: CreateNotificationPreferences :one
INSERT INTO notifications.notification_preferences (
  user_id, push, email, marketing,
  quiet_hours_enabled, quiet_hours_start, quiet_hours_end, timezone
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpsertNotificationPreferences :one
INSERT INTO notifications.notification_preferences (
  user_id, push, email, marketing,
  quiet_hours_enabled, quiet_hours_start, quiet_hours_end, timezone
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
  push = EXCLUDED.push,
  email = EXCLUDED.email,
  marketing = EXCLUDED.marketing,
  quiet_hours_enabled = EXCLUDED.quiet_hours_enabled,
  quiet_hours_start = EXCLUDED.quiet_hours_start,
  quiet_hours_end = EXCLUDED.quiet_hours_end,
  timezone = EXCLUDED.timezone
RETURNING *;
