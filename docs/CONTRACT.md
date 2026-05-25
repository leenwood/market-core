# market-core — Product Catalog Service

Бэкенд-микросервис каталога товаров для маркетплейса. Предоставляет CRUD товаров и категорий,
полнотекстовый + нечёткий поиск на PostgreSQL, аналитику поисковых запросов и избранное.

**Стек:** Go 1.26 · Gin · PostgreSQL 16 (pgx v5) · Prometheus · OpenTelemetry

---

## Запуск

```
POST :8080              HTTP API
GET  /health            liveness
GET  /ready             readiness (ping DB)
GET  /metrics           Prometheus
GET  /swagger/index.html
```

---

## Контракты API

**Base path:** `/api/v1`

### Товары

| Метод | Путь | Тело запроса | Описание |
|---|---|---|---|
| `POST` | `/products` | `CreateProductRequest` | Создать товар |
| `GET` | `/products` | query params | Список с фильтрацией |
| `GET` | `/products/:id` | — | Получить по ID (инкрементирует view_count) |
| `PUT` | `/products/:id` | `UpdateProductRequest` | Частичное обновление |
| `DELETE` | `/products/:id` | — | Мягкое удаление |

**`CreateProductRequest`**
```json
{
  "name": "iPhone 15 Pro",
  "description": "...",
  "category_id": "uuid",
  "brand": "Apple",
  "price": 119990,
  "in_stock": true,
  "attributes": { "color": "black", "storage": "256GB" }
}
```

`name` и `category_id` — обязательные поля.

**`UpdateProductRequest`** — все поля опциональны (`*T`), отсутствующее поле не затрагивает запись.

**Query-параметры `GET /products`**

| Параметр | Тип | По умолчанию |
|---|---|---|
| `category_id` | UUID | — |
| `include_subcategory` | bool | `false` |
| `brand` | string | — |
| `min_price` / `max_price` | float | — |
| `in_stock` | bool | — |
| `sort_by` | `price` / `created_at` / `popularity` / `rating` | — |
| `sort_dir` | `asc` / `desc` | — |
| `page` / `page_size` | int | `1` / `20` (max `100`) |

**`ProductResponse`**
```json
{
  "id": "uuid",
  "name": "...",
  "description": "...",
  "category_id": "uuid",
  "brand": "...",
  "price": 119990,
  "rating": 4.8,
  "rating_count": 123,
  "in_stock": true,
  "attributes": {},
  "view_count": 4200,
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

Списочные эндпоинты возвращают обёртку:
```json
{
  "items": [...],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

---

### Категории

| Метод | Путь | Тело запроса | Описание |
|---|---|---|---|
| `POST` | `/categories` | `CreateCategoryRequest` | Создать категорию |
| `GET` | `/categories` | — | Полное дерево |
| `GET` | `/categories/:id` | — | Категория по ID |
| `DELETE` | `/categories/:id` | — | Удалить (ошибка если есть товары) |

**`CreateCategoryRequest`**
```json
{
  "name": "Смартфоны",
  "slug": "smartphones",
  "parent_id": "uuid",
  "sort_order": 0
}
```

`name` и `slug` — обязательные поля. `parent_id: null` — корневая категория.

**`CategoryResponse`** — возвращается рекурсивное дерево через поле `children: []CategoryResponse`.

---

### Поиск

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/search` | FTS + нечёткий поиск с фильтрами |
| `GET` | `/search/autocomplete` | Подсказки по префиксу |

**Query-параметры `GET /search`**

| Параметр | Тип | Описание |
|---|---|---|
| `q` | string | Поисковый запрос |
| `category_id` | UUID | Фильтр с раскрытием подкатегорий |
| `brand` | string | |
| `min_price` / `max_price` | float | |
| `in_stock` | bool | |
| `sort_by` | `relevance` / `price` / `created_at` / `popularity` | default: `relevance` |
| `sort_dir` | `asc` / `desc` | default: `desc` |
| `page` / `page_size` | int | default: `1` / `20` |

Заголовок `X-User-ID: <uuid>` — сохраняет запрос в историю поиска пользователя.

**`SearchResponse`**
```json
{
  "items": [...],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3,
  "query_id": "uuid"
}
```

`query_id` — идентификатор для трекинга кликов через аналитику.

**`GET /search/autocomplete?q=iph&limit=5`**
```json
{ "suggestions": ["iPhone 15", "iPhone 14", "..."] }
```

---

### Избранное

Все запросы требуют заголовка `X-User-ID: <uuid>`.

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/favorites` | Список избранных товаров пользователя |
| `POST` | `/favorites?product_id=<uuid>` | Добавить товар в избранное |
| `DELETE` | `/favorites?product_id=<uuid>` | Убрать товар из избранного |

---

### Аналитика

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/analytics/popular-queries?limit=10` | Топ поисковых запросов за 30 дней |
| `GET` | `/analytics/popular-products?limit=10` | Самые просматриваемые товары |

---

## Коды ответов

| Код | Ситуация |
|---|---|
| `200` | OK |
| `201` | Создан |
| `400` | Невалидный запрос / отсутствует обязательное поле |
| `404` | Ресурс не найден |
| `409` | Конфликт (slug уже занят, товар уже в избранном) |
| `500` | Внутренняя ошибка |

Тело ошибки: `{ "error": "описание" }`

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `DATABASE_DSN` | **обязательно** | PostgreSQL DSN |
| `HTTP_ADDR` | `:8080` | Адрес сервера |
| `HTTP_READ_TIMEOUT` | `15s` | |
| `HTTP_WRITE_TIMEOUT` | `15s` | |
| `HTTP_IDLE_TIMEOUT` | `60s` | |
| `HTTP_PPROF_ENABLED` | `false` | Включить `/debug/pprof/*` |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | `json` | `json` / `text` |
| `OTEL_ENABLED` | `false` | Включить трейсинг |
| `OTEL_EXPORTER` | `stdout` | `stdout` / `otlp` |
| `OTEL_ENDPOINT` | — | OTLP endpoint (только для `otlp`) |
| `OTEL_SERVICE_NAME` | `market-core` | Имя сервиса в трейсах |
