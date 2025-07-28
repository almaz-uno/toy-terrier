-- name: CreateSubscription :one
INSERT INTO subscriptions (
    user_id, category_id, is_active
) VALUES (
    $1, $2, $3
) ON CONFLICT (user_id, category_id)
DO UPDATE SET is_active = $3, updated_at = NOW()
RETURNING *;

-- name: GetUserSubscriptions :many
SELECT s.*, c.name as category_name, c.slug as category_slug
FROM subscriptions s
JOIN categories c ON s.category_id = c.id
WHERE s.user_id = $1 AND s.is_active = true
ORDER BY c.name;

-- name: GetSubscriptionByUserAndCategory :one
SELECT * FROM subscriptions
WHERE user_id = $1 AND category_id = $2;

-- name: UpdateSubscription :exec
UPDATE subscriptions
SET is_active = $2, updated_at = NOW()
WHERE user_id = $1 AND category_id = $3;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE user_id = $1 AND category_id = $2;

-- name: GetSubscribersByCategory :many
SELECT u.* FROM users u
JOIN subscriptions s ON u.id = s.user_id
WHERE s.category_id = $1 AND s.is_active = true AND u.is_active = true;

-- name: GetActiveSubscriptions :many
SELECT s.*, u.telegram_id, u.username, c.name as category_name
FROM subscriptions s
JOIN users u ON s.user_id = u.id
JOIN categories c ON s.category_id = c.id
WHERE s.is_active = true AND u.is_active = true
ORDER BY c.name, u.username;

-- name: CountSubscriptions :one
SELECT COUNT(*) FROM subscriptions WHERE is_active = true;

-- name: CountUserSubscriptions :one
SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND is_active = true;

-- name: DeactivateUserSubscriptions :exec
UPDATE subscriptions
SET is_active = false, updated_at = NOW()
WHERE user_id = $1;
