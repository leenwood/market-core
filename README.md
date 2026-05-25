# market-core

[🇷🇺 Русский](README.ru.md)

A production-oriented Go backend microservice for marketplace product catalogs.
Built to demonstrate real-world patterns in full-text search, hierarchical data,
dynamic filtering, and search analytics on top of PostgreSQL.

---

## What This Repository Demonstrates

- **PostgreSQL Full-Text Search** — `websearch_to_tsquery` over a weighted `tsvector` column auto-maintained by a trigger
- **Trigram fuzzy search** — `pg_trgm` similarity for typo-tolerant matching, combined with FTS in a single query
- **Relevance ranking** — blended score `ts_rank_cd × 0.7 + word_similarity × 0.3`
- **Hierarchical categories** — parent/child tree with recursive CTEs for subcategory expansion
- **JSONB dynamic attributes** — arbitrary product characteristics with GIN-indexed filtering
- **Search analytics** — query recording, click-through tracking, popular queries over a rolling 30-day window
- **Layered architecture** — strict `domain → port → usecase → infra → app` dependency graph enforced by `.go-arch-lint.yml`
- **Prometheus metrics** — request counters, latency histograms, search and view counters
- **Structured logging** — JSON via `log/slog` with `request_id` on every line
- **Swagger UI** — auto-generated from handler annotations, served at `/swagger/index.html`

---

## Architecture Overview

```
                    ┌────────────────────────────────────────┐
                    │              Gin HTTP Server            │
                    │  /api/v1/products   /api/v1/categories  │
                    │  /api/v1/search     /api/v1/favorites   │
                    │  /api/v1/analytics  /health  /metrics   │
                    │  /swagger/*                             │
                    └──────────────────┬─────────────────────┘
                                       │
                    ┌──────────────────▼─────────────────────┐
                    │              Use Cases                  │
                    │  ProductUC  CategoryUC  SearchUC        │
                    │  AnalyticsUC  FavoritesUC               │
                    └──────────────────┬─────────────────────┘
                                       │  port interfaces
                    ┌──────────────────▼─────────────────────┐
                    │          PostgreSQL 16 (pgx v5)         │
                    │                                         │
                    │  products         categories            │
                    │  ┌─────────────┐  ┌──────────────────┐ │
                    │  │ tsvector    │  │ parent_id (tree) │ │
                    │  │ GIN index   │  │ recursive CTE    │ │
                    │  │ JSONB attrs │  └──────────────────┘ │
                    │  │ GIN index   │                        │
                    │  │ trigram idx │  search_queries        │
                    │  └─────────────┘  search_clicks         │
                    │                   search_history        │
                    │                   favorites             │
                    └─────────────────────────────────────────┘

  Observability:
  Prometheus ← scrapes /metrics
  slog       → stdout (JSON)
  Swagger UI → /swagger/index.html
```

---

## Search Request Lifecycle

1. Client sends `GET /api/v1/search?q=iphone+15&brand=Apple&max_price=120000`
2. `RequestID` middleware assigns `X-Request-ID` UUID
3. `SearchHandler` parses query params into `dto.SearchRequest`
4. `SearchUseCase.Execute()` normalises page/sort defaults
5. `SearchRepo.Search()` builds dynamic parameterised SQL:
   - `websearch_to_tsquery('simple', $1)` matched against `search_vector`
   - `word_similarity($1, name) > 0.2` catches typos via trigram index
   - SQL filters appended per non-nil field (`brand ILIKE`, `price >=`, etc.)
   - `category_id IN (WITH RECURSIVE …)` expands to all subcategories
6. PostgreSQL executes with GIN + B-tree indexes, returns ranked rows
7. Results ordered by `ts_rank_cd × 0.7 + word_similarity × 0.3`
8. `SearchQuery` row inserted asynchronously (goroutine) for analytics
9. If `X-User-ID` header present — query appended to search history
10. Handler returns `dto.SearchResponse` with `query_id`, total, pages

---

## Full-Text Search and Ranking

### tsvector trigger

```sql
CREATE OR REPLACE FUNCTION products_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', COALESCE(NEW.name, '')),        'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(NEW.brand, '')),       'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

Name (weight A) ranks higher than description (B) and brand (C).

### Fuzzy fallback

When FTS produces no matches (e.g. severe typo), `word_similarity` from `pg_trgm`
kicks in with threshold `> 0.2`, so `"iphon"` still finds iPhone products.

### Indexes

| Index | Type | Purpose |
|---|---|---|
| `idx_products_search_vector` | GIN | FTS lookups |
| `idx_products_name_trgm` | GIN (trigram) | Fuzzy / autocomplete |
| `idx_products_description_trgm` | GIN (trigram) | Fuzzy on description |
| `idx_products_attributes` | GIN | JSONB attribute filters |
| `idx_products_price` | B-tree (partial) | Price range scans |
| `idx_products_category_id` | B-tree (partial) | Category filter |

---

## Analytics

Search queries are recorded to `search_queries` on every non-empty search.
Clicks are tracked via `POST /api/v1/analytics/click` linking a `search_query_id`
to a `product_id`.

```
GET /api/v1/analytics/popular-queries?limit=10   — top queries last 30 days
GET /api/v1/analytics/popular-products?limit=10  — most viewed products
```

Product view counts are incremented asynchronously on every `GET /products/:id`.

---

## Local Setup

**Prerequisites:** Docker, Docker Compose, Go 1.26+, [go-task](https://taskfile.dev/installation/)

```bash
# Install go-task (once)
go install github.com/go-task/task/v3/cmd/task@latest

# 1. Clone and enter
git clone https://github.com/leenwood/market-core
cd market-core

# 2. Copy env
cp .env.example .env

# 3. Start PostgreSQL
task docker:up

# 4. Apply migrations
task migrate:up

# 5. Run the server
task run
```

Services available locally:

| Service | URL |
|---|---|
| API server | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Health check | http://localhost:8080/health |
| Prometheus metrics | http://localhost:8080/metrics |

---

## API Examples

```bash
# Health check
curl http://localhost:8080/health

# Create a category
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Smartphones", "slug": "smartphones"}'

# Create a product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro Max",
    "description": "Apple flagship smartphone with titanium frame",
    "category_id": "<category-uuid>",
    "brand": "Apple",
    "price": 119990,
    "in_stock": true,
    "attributes": {"color": "black", "storage": "256GB"}
  }'

# Full-text search with filters
curl "http://localhost:8080/api/v1/search?q=iphone+15&brand=Apple&max_price=120000&sort_by=relevance"

# Fuzzy search (typo)
curl "http://localhost:8080/api/v1/search?q=iphon+pro"

# Autocomplete
curl "http://localhost:8080/api/v1/search/autocomplete?q=iph&limit=5"

# Add to favorites (X-User-ID required)
curl -X POST "http://localhost:8080/api/v1/favorites?product_id=<uuid>" \
  -H "X-User-ID: <user-uuid>"

# Popular search queries
curl "http://localhost:8080/api/v1/analytics/popular-queries?limit=10"
```

---

## Useful Task Commands

Run `task --list` to see all available commands.

| Command | Description |
|---|---|
| `task build` | Build server and migrate binaries |
| `task build:server` | Build server binary → `bin/server` |
| `task build:migrate` | Build migrate binary → `bin/migrate` |
| `task run` | Run HTTP server locally (reads `.env`) |
| `task test` | Run unit tests with race detector |
| `task test:integration` | Run integration tests (requires Docker) |
| `task test:cover` | Tests with HTML coverage report |
| `task lint` | Run golangci-lint |
| `task vet` | Run go vet |
| `task fmt` | Format code with gofmt + goimports |
| `task arch` | Check architectural dependency rules |
| `task arch:graph` | Generate arch graph → `docs/arch-graph.svg` |
| `task docker:up` | Start PostgreSQL container |
| `task docker:down` | Stop containers |
| `task docker:reset` | Stop containers and wipe volumes |
| `task docker:logs` | Stream container logs |
| `task migrate:up` | Apply all pending migrations |
| `task migrate:down` | Rollback last migration |
| `task migrate:status` | Print migration status |
| `task migrate:create -- <name>` | Create new migration file |
| `task docs` | Regenerate Swagger docs from annotations |
| `task deps` | Download and tidy Go dependencies |

---

## Testing

Unit tests cover use case logic with in-memory stubs for all port interfaces.

Integration tests use `testcontainers-go` to spin up a real PostgreSQL instance,
verifying the full flow: migration → product create → FTS search → filter → analytics.

```bash
task test
task test:integration
```

---

## Possible Production Improvements

- **Redis autocomplete** — move prefix suggestions to a Redis sorted set for sub-millisecond latency at scale
- **Elasticsearch / Typesense** — replace PostgreSQL FTS for multilingual search, synonyms, and faceted navigation
- **Outbox pattern** — write search analytics to an outbox table and relay asynchronously, eliminating the goroutine fire-and-forget
- **Rate limiting** — per-user request limits on search endpoints to prevent scraping
- **Category path (ltree)** — replace recursive CTE with PostgreSQL `ltree` extension for O(1) ancestor/descendant queries on deep trees
- **Recommendation engine** — collaborative filtering on `search_clicks` and `favorites` to power "similar products"
- **Cache layer** — Redis cache for popular search queries and category trees with TTL invalidation on writes
- **Audit log** — append-only table tracking every product status change for compliance and rollback
