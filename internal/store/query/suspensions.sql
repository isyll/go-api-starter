-- name: GetActiveSuspensionByUserID :one
SELECT * FROM auth.account_suspensions
WHERE user_id = $1 AND (is_permanent = TRUE OR suspended_until > now())
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSuspension :one
INSERT INTO auth.account_suspensions (
  user_id, reason, details, suspended_until, is_permanent
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeactivateActiveSuspensions :exec
UPDATE auth.account_suspensions
SET is_permanent = FALSE, suspended_until = now()
WHERE user_id = $1 AND (is_permanent = TRUE OR suspended_until > now());
