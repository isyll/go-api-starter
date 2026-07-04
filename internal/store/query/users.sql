-- name: GetUserByID :one
SELECT * FROM auth.users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM auth.users WHERE email = $1 AND deleted_at IS NULL;

-- name: ExistsUserByEmail :one
SELECT EXISTS (
  SELECT 1 FROM auth.users WHERE email = $1 AND deleted_at IS NULL
);

-- name: CreateUser :one
INSERT INTO auth.users (email, password_hash, first_name, last_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE auth.users SET last_login_at = now() WHERE id = $1;

-- name: UpdateUserPasswordHash :exec
UPDATE auth.users SET password_hash = $2 WHERE id = $1;

-- name: MarkUserEmailVerified :exec
UPDATE auth.users SET email_verified_at = now() WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE auth.users SET
  first_name = COALESCE(sqlc.narg('first_name'), first_name),
  last_name  = COALESCE(sqlc.narg('last_name'), last_name),
  bio        = COALESCE(sqlc.narg('bio'), bio),
  avatar     = COALESCE(sqlc.narg('avatar'), avatar)
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserStatus :exec
UPDATE auth.users SET status = $2 WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE auth.users SET role = $2 WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE auth.users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM auth.users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM auth.users WHERE deleted_at IS NULL;
