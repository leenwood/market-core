# market-core

Backend microservice for a marketplace product catalog. Provides product and category management, full-text search powered by PostgreSQL, filtering, analytics, favorites, and search history.

[Русская версия](README.ru.md)

---

## Features

- **Product CRUD** — create, read, update, soft-delete products
- **Category hierarchy** — tree structure with recursive subcategory queries
- **Product attributes** — brand, price, rating, stock status, arbitrary JSONB fields
- **Full-text search** — `websearch_to_tsquery` over `tsvector` column (auto-updated via trigger)
- **Fuzzy search** — trigram similarity via `pg_trgm`, handles typos
- **Combined ranking** — `ts_rank_cd × 0.7 + word_similarity × 0.3`
- **Filtering** — price range, brand, in-stock flag, category (incl. subcategories), dynamic JSONB filters
- **Sorting** — by relevance, price, creation date, popularity (view count)
- **Pagination** — page/page_size on all list endpoints
- **Autocomplete** — prefix match on product names and past queries
- **Search analytics** — query recording, popular queries (last 30 days), click-through tracking
- **Favorites** — per-user product lists
- **Search history** — per-user query history
- **Prometheus metrics** — request counters, latency histograms
- **Structured logging** — JSON via `log/slog`

---

## Tech stack

| Component      | Technology                        |
|---------------|-----------------------------------|
| Language       | Go 1.23                           |
| HTTP router    | [chi v5](https://github.com/go-chi/chi) |
| Database       | PostgreSQL 16                     |
| DB driver      | [pgx v5](https://github.com/jackc/pgx) |
| Migrations     | [goose v3](https://github.com/pressly/goose) |
| Metrics        | [Prometheus](https://github.com/prometheus/client_golang) |
| Containerization | Docker / Docker Compose         |
| Task runner    | [Task](https://taskfile.dev)      |

---

## Project structure

```
cmd/
├── server/         HTTP server entrypoint
└── migrate/        DB migration runner

internal/
├── core/
│   ├── domain/     Domain entities and sentinel errors
│   ├── dto/        Request / response data transfer objects
│   ├── port/       Repository interfaces (no infra deps)
│   ├── mapper/     domain ↔ dto conversions
│   └── usecase/    Business logic per domain (product, category, search, analytics, favorites)
├── infra/
│   └── storage/
│       └── postgres/   pgx implementations of all port interfaces
├── platform/
│   ├── logger/     Structured slog setup
│   └── metrics/    Prometheus counters and histograms
└── app/
    ├── http/       chi router, handlers, middleware
    ├── migrate/    goose runner wrapper
    └── service/    Bootstrap: wires everything and starts the server

migrations/         goose SQL migration files
```

---

## Getting started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- [Task](https://taskfile.dev/#/installation) (`brew install go-task`)

### 1. Configure environment

```bash
cp .env.example .env
# Edit .env if needed (defaults work with docker-compose)
```

### 2. Start the database

```bash
task docker:up
```

### 3. Apply migrations

```bash
task migrate:up
```

### 4. Run the server

```bash
task run
# Server starts on :8080
```

---

## API reference

Base path: `/api/v1`

### Products

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/products` | Create product |
| `GET` | `/products` | List products with filters and pagination |
| `GET` | `/products/{id}` | Get product by ID |
| `PUT` | `/products/{id}` | Update product |
| `DELETE` | `/products/{id}` | Soft-delete product |

**List / filter query params:**

| Param | Type | Description |
|-------|------|-------------|
| `category_id` | UUID | Filter by category |
| `include_subcategory` | bool | Include subcategories (default `false`) |
| `brand` | string | Brand name (partial match) |
| `min_price` | float | Minimum price |
| `max_price` | float | Maximum price |
| `in_stock` | bool | Stock availability |
| `sort_by` | string | `price` / `created_at` / `popularity` / `rating` |
| `sort_dir` | string | `asc` / `desc` |
| `page` | int | Page number (default `1`) |
| `page_size` | int | Items per page (default `20`, max `100`) |

### Categories

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/categories` | Create category |
| `GET` | `/categories` | Get full category tree |
| `GET` | `/categories/{id}` | Get category by ID |
| `DELETE` | `/categories/{id}` | Delete category |

### Search

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/search` | Full-text search with filters |
| `GET` | `/search/autocomplete` | Autocomplete suggestions |

**Search query params:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search query (FTS + fuzzy) |
| `category_id` | UUID | Filter by category (incl. subcategories) |
| `brand` | string | Brand filter |
| `min_price` / `max_price` | float | Price range |
| `in_stock` | bool | Stock filter |
| `sort_by` | string | `relevance` / `price` / `created_at` / `popularity` |
| `sort_dir` | string | `asc` / `desc` |
| `page` / `page_size` | int | Pagination |

Pass `X-User-ID: <uuid>` header to record search history.

### Favorites

Requires `X-User-ID: <uuid>` header on all requests.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/favorites` | List favorite products |
| `POST` | `/favorites?product_id={id}` | Add to favorites |
| `DELETE` | `/favorites?product_id={id}` | Remove from favorites |

### Analytics

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/analytics/popular-queries?limit=10` | Top search queries (last 30 days) |
| `GET` | `/analytics/popular-products?limit=10` | Most viewed products |

### System

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |

---

## Example scenarios

**Scenario 1 — text search with typos:**

```
GET /api/v1/search?q=iphone+15&sort_by=relevance&page=1&page_size=20
```

The service matches against the FTS index, falls back to trigram similarity for typo tolerance, and ranks results by `ts_rank_cd × 0.7 + word_similarity × 0.3`.

**Scenario 2 — combined filter + text search:**

```
GET /api/v1/search?q=pro+max&category_id=<uuid>&brand=Apple&max_price=120000&in_stock=true&sort_by=relevance
```

Full-text search is applied on top of SQL filters in a single query.

---

## Development

```bash
task test          # unit tests with race detector
task lint          # golangci-lint
task arch          # architectural dependency check
task fmt           # gofmt + goimports
task vet           # go vet
task test:cover    # coverage report → coverage.html
```

### Create a migration

```bash
task migrate:create -- add_product_tags
```

### Database access

```bash
docker compose exec postgres psql -U market -d market_core
```

---

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_DSN` | — | PostgreSQL connection string (**required**) |
| `HTTP_ADDR` | `:8080` | HTTP listen address |
| `LOG_LEVEL` | `info` | Log level: `debug` / `info` / `warn` / `error` |
| `MIGRATION_DIR` | `migrations` | Path to goose migration files |

---

## License

MIT
