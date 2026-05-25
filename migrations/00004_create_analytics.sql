-- +goose Up
CREATE TABLE search_queries (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID,
    query         TEXT        NOT NULL,
    results_count INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE search_clicks (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    search_query_id UUID        NOT NULL REFERENCES search_queries(id) ON DELETE CASCADE,
    product_id      UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE search_history (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    query      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_queries_user_id   ON search_queries(user_id);
CREATE INDEX idx_search_queries_created_at ON search_queries(created_at DESC);
CREATE INDEX idx_search_queries_query     ON search_queries(query);
CREATE INDEX idx_search_history_user_id   ON search_history(user_id);
CREATE INDEX idx_search_history_created_at ON search_history(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS search_clicks;
DROP TABLE IF EXISTS search_queries;
