-- name: CreateLoginAttempt :exec
INSERT INTO auth.login_attempts (
  email, user_id, channel, outcome, remaining,
  ip_address, user_agent, device_id, request_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: DeleteLoginAttemptsBefore :execrows
DELETE FROM auth.login_attempts WHERE created_at < $1;
