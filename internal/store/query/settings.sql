-- name: GetUserSettings :one
SELECT * FROM auth.user_settings WHERE user_id = $1;

-- name: CreateUserSettings :exec
INSERT INTO auth.user_settings (user_id, settings) VALUES ($1, $2);

-- name: UpdateUserSettings :exec
UPDATE auth.user_settings SET settings = $2 WHERE user_id = $1;
