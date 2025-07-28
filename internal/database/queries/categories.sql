-- name: GetActiveCategories :many
SELECT * FROM categories WHERE is_active = true ORDER BY name;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: GetCategoriesBySourceType :many
SELECT * FROM categories WHERE source_type = $1 AND is_active = true ORDER BY name;

-- name: CreateCategory :one
INSERT INTO categories (
    name, slug, source_type, source_url, is_active
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, source_url = $3, is_active = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories WHERE is_active = true;
