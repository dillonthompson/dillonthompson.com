-- name: GetPublishedExperiences :many
SELECT id, company, role, start_date, end_date, description, tech_stack, security_highlights, metadata, created_at
FROM experiences
WHERE is_published = true
ORDER BY sort_order ASC, start_date DESC;

-- name: GetProfileSection :one
SELECT content FROM profile WHERE key = $1;

-- name: GetFullProfile :many
SELECT key, content FROM profile ORDER BY key;
