-- name: UpsertFCMToken :one
INSERT INTO auth.fcm_tokens (
  user_id, device_id, token, platform, app_version, is_active
) VALUES ($1, $2, $3, $4, $5, TRUE)
ON CONFLICT (user_id, device_id) DO UPDATE SET
  token = EXCLUDED.token,
  platform = EXCLUDED.platform,
  app_version = EXCLUDED.app_version,
  is_active = TRUE,
  updated_at = now()
RETURNING *;

-- name: ListFCMTokensByUserID :many
SELECT * FROM auth.fcm_tokens WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetFCMTokenByUserAndDevice :one
SELECT * FROM auth.fcm_tokens WHERE user_id = $1 AND device_id = $2;

-- name: DeleteFCMTokenByDevice :exec
DELETE FROM auth.fcm_tokens WHERE user_id = $1 AND device_id = $2;

-- name: ListActiveFCMTokensByUserID :many
SELECT * FROM auth.fcm_tokens WHERE user_id = $1 AND is_active = TRUE;

-- name: DeactivateFCMToken :exec
UPDATE auth.fcm_tokens SET is_active = FALSE WHERE id = $1;

-- name: TouchFCMTokenLastUsed :exec
UPDATE auth.fcm_tokens SET last_used_at = now() WHERE id = $1;
