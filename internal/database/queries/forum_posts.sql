-- name: CreateForumPost :one
INSERT INTO forum_topics (
    category_id, title, author, source_topic_id, source_url,
    content, reply_count, view_count, last_post_time,
    content_hash, has_images, image_urls, is_notified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, false
) RETURNING *;

-- name: GetForumPostBySourceID :one
SELECT * FROM forum_topics
WHERE source_topic_id = $1 AND category_id = $2;

-- name: GetForumPostByURL :one
SELECT * FROM forum_topics WHERE source_url = $1;

-- name: GetForumPostByContentHash :one
SELECT * FROM forum_topics WHERE content_hash = $1;

-- name: GetRecentForumPosts :many
SELECT ft.*, c.name as category_name
FROM forum_topics ft
JOIN categories c ON ft.category_id = c.id
WHERE ft.created_at > $1
ORDER BY ft.created_at DESC
LIMIT $2;

-- name: GetUnnotifiedPosts :many
SELECT * FROM forum_topics
WHERE is_notified = false
ORDER BY created_at ASC;

-- name: MarkPostAsNotified :exec
UPDATE forum_topics
SET is_notified = true, updated_at = NOW()
WHERE id = $1;

-- name: GetPostsByCategory :many
SELECT * FROM forum_topics
WHERE category_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPostsByCategory :one
SELECT COUNT(*) FROM forum_topics WHERE category_id = $1;

-- name: CountTotalPosts :one
SELECT COUNT(*) FROM forum_topics;

-- name: CountPostsLast24h :one
SELECT COUNT(*) FROM forum_topics
WHERE created_at > NOW() - INTERVAL '24 hours';

-- name: GetLastScrapingTime :one
SELECT MAX(created_at) FROM forum_topics;
