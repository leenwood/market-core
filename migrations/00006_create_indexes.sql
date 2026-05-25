-- +goose Up

-- GIN index for FTS
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

-- GIN index for JSONB attributes
CREATE INDEX idx_products_attributes ON products USING GIN(attributes);

-- Trigram indexes for fuzzy search
CREATE INDEX idx_products_name_trgm        ON products USING GIN(name gin_trgm_ops);
CREATE INDEX idx_products_description_trgm ON products USING GIN(description gin_trgm_ops);
CREATE INDEX idx_products_brand_trgm       ON products USING GIN(brand gin_trgm_ops);

-- B-tree indexes for filtering and sorting
CREATE INDEX idx_products_category_id ON products(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_brand       ON products(brand)       WHERE deleted_at IS NULL;
CREATE INDEX idx_products_price       ON products(price)       WHERE deleted_at IS NULL;
CREATE INDEX idx_products_in_stock    ON products(in_stock)    WHERE deleted_at IS NULL;
CREATE INDEX idx_products_view_count  ON products(view_count DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_created_at  ON products(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_deleted_at  ON products(deleted_at);

-- Trigram index for autocomplete on search_queries
CREATE INDEX idx_search_queries_query_trgm ON search_queries USING GIN(query gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_search_queries_query_trgm;
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_view_count;
DROP INDEX IF EXISTS idx_products_in_stock;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_brand;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_brand_trgm;
DROP INDEX IF EXISTS idx_products_description_trgm;
DROP INDEX IF EXISTS idx_products_name_trgm;
DROP INDEX IF EXISTS idx_products_attributes;
DROP INDEX IF EXISTS idx_products_search_vector;
