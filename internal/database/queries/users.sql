-- name: CreateUser :one
INSERT INTO users (
    telegram_id, username, first_name, last_name, language_code, is_active, is_admin
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserActivity :exec
UPDATE users
SET last_activity = NOW(), updated_at = NOW()
WHERE telegram_id = $1;

-- name: UpdateUser :one
UPDATE users
SET username = $2, first_name = $3, last_name = $4, language_code = $5, updated_at = NOW()
WHERE telegram_id = $1
RETURNING *;

-- name: SetUserActive :exec
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE telegram_id = $1;

-- name: GetActiveUsers :many
SELECT * FROM users WHERE is_active = true ORDER BY created_at DESC;

-- name: GetAllUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: GetAdminUsers :many
SELECT * FROM users WHERE is_admin = true AND is_active = true;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountActiveUsers :one
SELECT COUNT(*) FROM users WHERE is_active = true;

-- name: GetRecentUsers :many
SELECT * FROM users
WHERE is_active = true
ORDER BY last_activity DESC
LIMIT $1;
