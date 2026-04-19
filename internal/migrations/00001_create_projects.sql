-- +goose Up
CREATE TABLE IF NOT EXISTS experiences (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company             TEXT NOT NULL,
    role                TEXT NOT NULL,
    start_date          DATE NOT NULL,
    end_date            DATE,
    description         JSONB NOT NULL DEFAULT '[]'::jsonb,
    tech_stack          JSONB NOT NULL DEFAULT '[]'::jsonb,
    security_highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_published        BOOLEAN NOT NULL DEFAULT false,
    sort_order          INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_experiences_published ON experiences (is_published, sort_order) WHERE is_published = true;
CREATE INDEX idx_experiences_tech_stack ON experiences USING GIN (tech_stack);

-- +goose Down
DROP TABLE IF EXISTS experiences;
