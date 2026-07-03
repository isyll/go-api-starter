-- name: GetUserByID :one
SELECT * FROM auth.users WHERE id = $1 AND deleted_at IS NULL;
