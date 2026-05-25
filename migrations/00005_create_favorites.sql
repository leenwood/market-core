-- +goose Up
CREATE TABLE favorites (
    user_id    UUID        NOT NULL,
    product_id UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, product_id)
);

CREATE INDEX idx_favorites_user_id ON favorites(user_id);

-- +goose Down
DROP TABLE IF EXISTS favorites;
