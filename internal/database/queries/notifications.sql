-- name: CreateNotification :one
INSERT INTO notifications (
    user_id, topic_id, message_text, status, retry_count
) VALUES (
    $1, $2, $3, 'pending', 0
) RETURNING *;

-- name: GetPendingNotifications :many
SELECT * FROM notifications
WHERE status = 'pending' OR (status = 'rate_limited' AND next_retry_at <= NOW())
ORDER BY created_at ASC
LIMIT $1;

-- name: UpdateNotificationStatus :exec
UPDATE notifications
SET status = $2, telegram_message_id = $3, sent_at = $4, error_message = $5
WHERE id = $1;

-- name: UpdateNotificationRetry :exec
UPDATE notifications
SET status = $2, retry_count = retry_count + 1,
    next_retry_at = $3, error_message = $4
WHERE id = $1;

-- name: GetNotificationsByUser :many
SELECT n.*, ft.title as topic_title, ft.source_url as topic_url
FROM notifications n
JOIN forum_topics ft ON n.topic_id = ft.id
WHERE n.user_id = $1
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFailedNotifications :many
SELECT * FROM notifications
WHERE status = 'failed' AND retry_count < $1
ORDER BY created_at ASC;

-- name: CountNotificationsByStatus :one
SELECT COUNT(*) FROM notifications WHERE status = $1;

-- name: CountUserNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < $1 AND status IN ('sent', 'failed');

-- name: GetNotificationStats :one
SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE status = 'pending') as pending,
    COUNT(*) FILTER (WHERE status = 'sent') as sent,
    COUNT(*) FILTER (WHERE status = 'failed') as failed,
    COUNT(*) FILTER (WHERE status = 'rate_limited') as rate_limited
FROM notifications;
