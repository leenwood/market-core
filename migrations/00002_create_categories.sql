-- +goose Up
CREATE TABLE categories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(255) NOT NULL UNIQUE,
    parent_id  UUID        REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT         NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);

-- +goose Down
DROP TABLE IF EXISTS categories;
