-- name: CreateDeviceSession :one
INSERT INTO auth.device_sessions (
  user_id, platform, manufacturer, model, device_id,
  name, ip_address, user_agent, last_activity
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
RETURNING *;

-- name: GetDeviceSessionByID :one
SELECT * FROM auth.device_sessions WHERE id = $1;

-- name: GetActiveDeviceSessionByUserAndDevice :one
SELECT * FROM auth.device_sessions
WHERE user_id = $1 AND device_id = $2 AND revoked_at IS NULL;

-- name: RevokeDeviceSession :one
UPDATE auth.device_sessions
SET revoked_at = now(), revoked_reason = $2
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: UpdateDeviceSessionActivity :exec
UPDATE auth.device_sessions
SET last_activity = now()
WHERE id = $1 AND revoked_at IS NULL;

-- name: CountActiveDevicesByUser :one
SELECT count(*) FROM auth.device_sessions
WHERE user_id = $1 AND revoked_at IS NULL AND last_activity > $2;

-- name: ListActiveDevicesByUser :many
SELECT * FROM auth.device_sessions
WHERE user_id = $1 AND revoked_at IS NULL AND last_activity > $2
ORDER BY last_activity DESC;

-- name: RevokeAllDeviceSessionsByUser :exec
UPDATE auth.device_sessions
SET revoked_at = now(), revoked_reason = $2
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: RevokeActiveDeviceSessionsByDeviceID :exec
UPDATE auth.device_sessions
SET revoked_at = now(), revoked_reason = $2
WHERE device_id = $1 AND revoked_at IS NULL;
