-- name: GetPublishedExperiences :many
SELECT id, company, role, start_date, end_date, description, tech_stack, security_highlights, metadata, created_at
FROM experiences
WHERE is_published = true
ORDER BY sort_order ASC, start_date DESC;

-- name: GetProfileSection :one
SELECT content FROM profile WHERE key = $1;

-- name: GetFullProfile :many
SELECT key, content FROM profile ORDER BY key;

-- name: ListPublishedPosts :many
SELECT id, slug, title, description, tags, published_at, updated_at
FROM posts
WHERE status = 'published'
ORDER BY published_at DESC;

-- name: GetPublishedPostBySlug :one
SELECT id, slug, title, description, body_md, tags, status, published_at, created_at, updated_at
FROM posts
WHERE slug = $1 AND status = 'published';

-- name: ListAllPosts :many
SELECT id, slug, title, description, tags, status, published_at, created_at, updated_at
FROM posts
ORDER BY updated_at DESC;

-- name: GetPostByID :one
SELECT id, slug, title, description, body_md, tags, status, published_at, created_at, updated_at
FROM posts
WHERE id = $1;

-- name: CreatePost :one
INSERT INTO posts (slug, title, description, body_md, tags, status, published_at)
VALUES (
    sqlc.arg(slug), sqlc.arg(title), sqlc.arg(description), sqlc.arg(body_md), sqlc.arg(tags), sqlc.arg(status),
    CASE WHEN sqlc.arg(status) = 'published' THEN now() END
)
RETURNING id, slug, title, description, body_md, tags, status, published_at, created_at, updated_at;

-- name: UpdatePost :one
UPDATE posts
SET slug = sqlc.arg(slug),
    title = sqlc.arg(title),
    description = sqlc.arg(description),
    body_md = sqlc.arg(body_md),
    tags = sqlc.arg(tags),
    status = sqlc.arg(status),
    published_at = CASE
        WHEN sqlc.arg(status) = 'published' AND published_at IS NULL THEN now()
        ELSE published_at
    END,
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, slug, title, description, body_md, tags, status, published_at, created_at, updated_at;

-- name: DeletePost :execrows
DELETE FROM posts WHERE id = $1;
