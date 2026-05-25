-- +goose Up
CREATE TABLE products (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(500)   NOT NULL,
    description  TEXT           NOT NULL DEFAULT '',
    category_id  UUID           NOT NULL REFERENCES categories(id),
    brand        VARCHAR(255)   NOT NULL DEFAULT '',
    price        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    rating       NUMERIC(3, 2)  NOT NULL DEFAULT 0,
    rating_count INT            NOT NULL DEFAULT 0,
    in_stock     BOOLEAN        NOT NULL DEFAULT TRUE,
    attributes   JSONB          NOT NULL DEFAULT '{}',
    search_vector TSVECTOR,
    view_count   BIGINT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE OR REPLACE FUNCTION products_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.brand, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_search_vector
    BEFORE INSERT OR UPDATE OF name, description, brand
    ON products
    FOR EACH ROW
    EXECUTE FUNCTION products_search_vector_update();

-- +goose Down
DROP TRIGGER IF EXISTS trg_products_search_vector ON products;
DROP FUNCTION IF EXISTS products_search_vector_update();
DROP TABLE IF EXISTS products;
