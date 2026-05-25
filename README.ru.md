# market-core

[English](README.md)

Производственно-ориентированный Go-бэкенд микросервис для каталога товаров маркетплейса.
Создан для демонстрации реальных паттернов: полнотекстовый поиск, иерархические данные,
динамическая фильтрация и поисковая аналитика на базе PostgreSQL.

---

## Что демонстрирует репозиторий

- **Полнотекстовый поиск PostgreSQL** — `websearch_to_tsquery` по взвешенной колонке `tsvector`, автоматически обновляемой триггером
- **Нечёткий поиск по триграммам** — `pg_trgm` сходство для устойчивого к опечаткам поиска, объединённого с FTS в одном запросе
- **Ранжирование по релевантности** — смешанная оценка `ts_rank_cd × 0.7 + word_similarity × 0.3`
- **Иерархические категории** — дерево родитель/потомок с рекурсивными CTE для разворачивания подкатегорий
- **Динамические атрибуты JSONB** — произвольные характеристики товаров с GIN-индексированной фильтрацией
- **Аналитика поиска** — запись запросов, отслеживание кликов, популярные запросы за скользящие 30 дней
- **Слоистая архитектура** — строгий граф зависимостей `domain → port → usecase → infra → app`, контролируемый `.go-arch-lint.yml`
- **Метрики Prometheus** — счётчики запросов, гистограммы задержек, счётчики поиска и просмотров
- **Структурированные логи** — JSON через `log/slog` с `request_id` в каждой строке
- **Swagger UI** — автогенерация из аннотаций хендлеров, доступен по `/swagger/index.html`

---

## Обзор архитектуры

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

  Наблюдаемость:
  Prometheus ← scrapes /metrics
  slog       → stdout (JSON)
  Swagger UI → /swagger/index.html
```

---

## Жизненный цикл поискового запроса

1. Клиент отправляет `GET /api/v1/search?q=iphone+15&brand=Apple&max_price=120000`
2. Middleware `RequestID` присваивает UUID `X-Request-ID`
3. `SearchHandler` парсит параметры запроса в `dto.SearchRequest`
4. `SearchUseCase.Execute()` нормализует значения page/sort по умолчанию
5. `SearchRepo.Search()` строит динамический параметризованный SQL:
   - `websearch_to_tsquery('simple', $1)` сопоставляется со `search_vector`
   - `word_similarity($1, name) > 0.2` ловит опечатки через триграммный индекс
   - SQL-фильтры добавляются для каждого ненулевого поля (`brand ILIKE`, `price >=` и т.д.)
   - `category_id IN (WITH RECURSIVE …)` разворачивается до всех подкатегорий
6. PostgreSQL выполняет запрос с GIN + B-tree индексами, возвращает отранжированные строки
7. Результаты упорядочиваются по `ts_rank_cd × 0.7 + word_similarity × 0.3`
8. Строка `SearchQuery` асинхронно вставляется (горутина) для аналитики
9. Если присутствует заголовок `X-User-ID` — запрос добавляется в историю поиска
10. Хендлер возвращает `dto.SearchResponse` с `query_id`, total, pages

---

## Полнотекстовый поиск и ранжирование

### Триггер tsvector

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

Название (вес A) ранжируется выше описания (B) и бренда (C).

### Нечёткий fallback

Когда FTS не даёт результатов (например, сильная опечатка), `word_similarity` из `pg_trgm`
срабатывает с порогом `> 0.2`, поэтому `"iphon"` всё равно находит iPhone-товары.

### Индексы

| Индекс | Тип | Назначение |
|---|---|---|
| `idx_products_search_vector` | GIN | FTS-поиск |
| `idx_products_name_trgm` | GIN (trigram) | Нечёткий поиск / автодополнение |
| `idx_products_description_trgm` | GIN (trigram) | Нечёткий поиск по описанию |
| `idx_products_attributes` | GIN | Фильтры по JSONB-атрибутам |
| `idx_products_price` | B-tree (partial) | Сканирование диапазона цен |
| `idx_products_category_id` | B-tree (partial) | Фильтр по категории |

---

## Аналитика

Поисковые запросы записываются в `search_queries` при каждом непустом поиске.
Клики отслеживаются через `POST /api/v1/analytics/click`, связывая `search_query_id`
с `product_id`.

```
GET /api/v1/analytics/popular-queries?limit=10   — топ запросов за последние 30 дней
GET /api/v1/analytics/popular-products?limit=10  — самые просматриваемые товары
```

Счётчики просмотров товаров инкрементируются асинхронно при каждом `GET /products/:id`.

---

## Локальная установка

**Требования:** Docker, Docker Compose, Go 1.26+, [go-task](https://taskfile.dev/installation/)

```bash
# Установка go-task (один раз)
go install github.com/go-task/task/v3/cmd/task@latest

# 1. Клонирование и переход в директорию
git clone https://github.com/leenwood/market-core
cd market-core

# 2. Копирование .env
cp .env.example .env

# 3. Запуск PostgreSQL
task docker:up

# 4. Применение миграций
task migrate:up

# 5. Запуск сервера
task run
```

Сервисы, доступные локально:

| Сервис | URL |
|---|---|
| API-сервер | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Health check | http://localhost:8080/health |
| Метрики Prometheus | http://localhost:8080/metrics |

---

## Примеры API

```bash
# Health check
curl http://localhost:8080/health

# Создать категорию
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Смартфоны", "slug": "smartphones"}'

# Создать товар
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro Max",
    "description": "Флагманский смартфон Apple с титановым корпусом",
    "category_id": "<category-uuid>",
    "brand": "Apple",
    "price": 119990,
    "in_stock": true,
    "attributes": {"color": "black", "storage": "256GB"}
  }'

# Полнотекстовый поиск с фильтрами
curl "http://localhost:8080/api/v1/search?q=iphone+15&brand=Apple&max_price=120000&sort_by=relevance"

# Нечёткий поиск (опечатка)
curl "http://localhost:8080/api/v1/search?q=iphon+pro"

# Автодополнение
curl "http://localhost:8080/api/v1/search/autocomplete?q=iph&limit=5"

# Добавить в избранное (требуется X-User-ID)
curl -X POST "http://localhost:8080/api/v1/favorites?product_id=<uuid>" \
  -H "X-User-ID: <user-uuid>"

# Популярные поисковые запросы
curl "http://localhost:8080/api/v1/analytics/popular-queries?limit=10"
```

---

## Команды Task

Выполните `task --list` для просмотра всех доступных команд.

| Команда | Описание |
|---|---|
| `task build` | Собрать бинарники server и migrate |
| `task build:server` | Собрать бинарник сервера → `bin/server` |
| `task build:migrate` | Собрать бинарник миграций → `bin/migrate` |
| `task run` | Запустить HTTP-сервер локально (читает `.env`) |
| `task test` | Запустить юнит-тесты с race detector |
| `task test:integration` | Запустить интеграционные тесты (требует Docker) |
| `task test:cover` | Тесты с HTML-отчётом о покрытии |
| `task lint` | Запустить golangci-lint |
| `task vet` | Запустить go vet |
| `task fmt` | Форматирование кода через gofmt + goimports |
| `task arch` | Проверить архитектурные правила зависимостей |
| `task arch:graph` | Сгенерировать граф архитектуры → `docs/arch-graph.svg` |
| `task docker:up` | Запустить контейнер PostgreSQL |
| `task docker:down` | Остановить контейнеры |
| `task docker:reset` | Остановить контейнеры и очистить тома |
| `task docker:logs` | Стримить логи контейнеров |
| `task migrate:up` | Применить все ожидающие миграции |
| `task migrate:down` | Откатить последнюю миграцию |
| `task migrate:status` | Показать статус миграций |
| `task migrate:create -- <name>` | Создать новый файл миграции |
| `task docs` | Перегенерировать Swagger-документацию из аннотаций |
| `task deps` | Скачать и привести в порядок Go-зависимости |

---

## Тестирование

Юнит-тесты покрывают логику use case с in-memory стабами для всех port-интерфейсов.

Интеграционные тесты используют `testcontainers-go` для запуска реального экземпляра PostgreSQL,
проверяя полный поток: миграция → создание товара → FTS-поиск → фильтрация → аналитика.

```bash
task test
task test:integration
```

---

## Возможные улучшения для продакшена

- **Redis-автодополнение** — вынести подсказки в отсортированное множество Redis для задержки менее миллисекунды при высокой нагрузке
- **Elasticsearch / Typesense** — заменить PostgreSQL FTS для многоязычного поиска, синонимов и фасетной навигации
- **Паттерн Outbox** — записывать поисковую аналитику в outbox-таблицу и асинхронно реплицировать, устранив горутины fire-and-forget
- **Rate limiting** — лимиты запросов по пользователю на поисковых эндпоинтах против скрейпинга
- **Category path (ltree)** — заменить рекурсивный CTE расширением PostgreSQL `ltree` для O(1) запросов предков/потомков на глубоких деревьях
- **Движок рекомендаций** — коллаборативная фильтрация по `search_clicks` и `favorites` для функционала «похожие товары»
- **Кэш-слой** — Redis-кеш для популярных поисковых запросов и деревьев категорий с TTL-инвалидацией при записи
- **Audit log** — append-only таблица с записью каждого изменения статуса товара для соответствия требованиям и возможности отката
