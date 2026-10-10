-- +goose Up
-- Blog posts. Body is stored as Markdown and rendered (and sanitized) by the Go
-- API at request time. published_at is stamped the first time a post moves to
-- 'published' and then left alone, so editing a live post doesn't reorder the
-- index or change its RSS date.
CREATE TABLE IF NOT EXISTS posts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    body_md      TEXT NOT NULL DEFAULT '',
    tags         TEXT[] NOT NULL DEFAULT '{}',
    status       TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT posts_published_has_date CHECK (status <> 'published' OR published_at IS NOT NULL)
);

CREATE INDEX idx_posts_published ON posts (published_at DESC) WHERE status = 'published';

-- +goose Down
DROP TABLE IF EXISTS posts;
