-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_tokens (
  session_id, token_hash, token_prefix, token_family, expires_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListRefreshTokensByPrefix :many
SELECT * FROM auth.refresh_tokens WHERE token_prefix = $1;

-- name: RevokeRefreshTokenByHash :exec
UPDATE auth.refresh_tokens
SET revoked_at = now(), revoked_reason = $2
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokensBySession :exec
UPDATE auth.refresh_tokens
SET revoked_at = now(), revoked_reason = $2
WHERE session_id = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokensByFamily :exec
UPDATE auth.refresh_tokens
SET revoked_at = now(), revoked_reason = $2
WHERE token_family = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE FROM auth.refresh_tokens
WHERE expires_at < now() AND revoked_at IS NULL;

-- name: DeleteStaleRefreshTokens :execrows
DELETE FROM auth.refresh_tokens
WHERE expires_at < $1
   OR (revoked_at IS NOT NULL AND revoked_at < $1);
