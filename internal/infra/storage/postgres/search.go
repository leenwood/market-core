package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRepo struct {
	db *pgxpool.Pool
}

func NewSearchRepo(db *pgxpool.Pool) *SearchRepo {
	return &SearchRepo{db: db}
}

func (r *SearchRepo) Search(ctx context.Context, req dto.SearchRequest) ([]*domain.Product, int64, error) {
	conds := []string{"p.deleted_at IS NULL"}
	args := []any{}
	n := 1

	rankExpr := "0.0::float8"
	if req.Query != "" {
		conds = append(conds, fmt.Sprintf(
			`(p.search_vector @@ websearch_to_tsquery('simple', $%d) OR word_similarity($%d, p.name) > 0.2)`,
			n, n,
		))
		rankExpr = fmt.Sprintf(
			`COALESCE(ts_rank_cd(p.search_vector, websearch_to_tsquery('simple', $%d)), 0)*0.7 + COALESCE(word_similarity($%d, p.name), 0)*0.3`,
			n, n,
		)
		args = append(args, req.Query)
		n++
	}

	if req.CategoryID != nil {
		conds = append(conds, fmt.Sprintf(`p.category_id IN (
			WITH RECURSIVE cats AS (
				SELECT id FROM categories WHERE id=$%d
				UNION ALL
				SELECT c.id FROM categories c INNER JOIN cats ON c.parent_id=cats.id
			) SELECT id FROM cats)`, n))
		args = append(args, *req.CategoryID)
		n++
	}
	if req.Brand != nil {
		conds = append(conds, fmt.Sprintf("p.brand ILIKE $%d", n))
		args = append(args, "%"+*req.Brand+"%")
		n++
	}
	if req.MinPrice != nil {
		conds = append(conds, fmt.Sprintf("p.price>=$%d", n))
		args = append(args, *req.MinPrice)
		n++
	}
	if req.MaxPrice != nil {
		conds = append(conds, fmt.Sprintf("p.price<=$%d", n))
		args = append(args, *req.MaxPrice)
		n++
	}
	if req.InStock != nil {
		conds = append(conds, fmt.Sprintf("p.in_stock=$%d", n))
		args = append(args, *req.InStock)
		n++
	}
	for k, v := range req.Attributes {
		conds = append(conds, fmt.Sprintf("p.attributes->>'%s'=$%d", sanitizeKey(k), n))
		args = append(args, fmt.Sprintf("%v", v))
		n++
	}

	where := "WHERE " + strings.Join(conds, " AND ")
	orderBy := buildSearchOrderBy(req.SortBy, req.SortDir, rankExpr)
	offset := (req.Page - 1) * req.PageSize

	countArgs := make([]any, len(args))
	copy(countArgs, args)
	var total int64
	if err := r.db.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM products p %s`, where),
		countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, req.PageSize, offset)
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.category_id, p.brand, p.price, p.rating, p.rating_count,
		       p.in_stock, p.attributes, p.view_count, p.created_at, p.updated_at, p.deleted_at,
		       (%s) AS rank
		FROM products p
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		rankExpr, where, orderBy, n, n+1,
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		var attrsRaw []byte
		var rank float64
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.CategoryID, &p.Brand, &p.Price,
			&p.Rating, &p.RatingCount, &p.InStock, &attrsRaw, &p.ViewCount,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &rank,
		); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
			p.Attributes = map[string]any{}
		}
		products = append(products, &p)
	}
	return products, total, rows.Err()
}

func buildSearchOrderBy(sortBy, sortDir, rankExpr string) string {
	dir := "DESC"
	if strings.ToLower(sortDir) == "asc" {
		dir = "ASC"
	}
	switch sortBy {
	case "price":
		return fmt.Sprintf("ORDER BY p.price %s", dir)
	case "created_at":
		return fmt.Sprintf("ORDER BY p.created_at %s", dir)
	case "popularity":
		return "ORDER BY p.view_count DESC"
	default:
		return fmt.Sprintf("ORDER BY (%s) DESC, p.created_at DESC", rankExpr)
	}
}

func (r *SearchRepo) Autocomplete(ctx context.Context, prefix string, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT name FROM (
			SELECT name FROM products
			WHERE name ILIKE $1 AND deleted_at IS NULL
			ORDER BY view_count DESC
			LIMIT $2
		) sub
		UNION
		SELECT DISTINCT query FROM (
			SELECT query FROM search_queries
			WHERE query ILIKE $1
			GROUP BY query
			ORDER BY COUNT(*) DESC
			LIMIT $2
		) sub2
		LIMIT $2`,
		prefix+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}
	return suggestions, rows.Err()
}

func (r *SearchRepo) RecordQuery(ctx context.Context, q *domain.SearchQuery) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO search_queries (id, user_id, query, results_count, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		q.ID, q.UserID, q.Query, q.ResultsCount, q.CreatedAt,
	)
	if err == nil && q.UserID != nil {
		_, _ = r.db.Exec(ctx, `
			INSERT INTO search_history (id, user_id, query, created_at)
			VALUES ($1, $2, $3, $4)`,
			uuid.New(), *q.UserID, q.Query, time.Now(),
		)
	}
	return err
}

func (r *SearchRepo) GetPopularQueries(ctx context.Context, limit int) ([]domain.PopularQuery, error) {
	rows, err := r.db.Query(ctx, `
		SELECT query, COUNT(*) AS cnt
		FROM search_queries
		WHERE created_at > NOW() - INTERVAL '30 days'
		GROUP BY query
		ORDER BY cnt DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PopularQuery
	for rows.Next() {
		var q domain.PopularQuery
		if err := rows.Scan(&q.Query, &q.Count); err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *SearchRepo) GetSearchHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.SearchHistory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, query, created_at
		FROM search_history WHERE user_id=$1
		ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.SearchHistory
	for rows.Next() {
		var h domain.SearchHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.Query, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, rows.Err()
}

func (r *SearchRepo) ClearSearchHistory(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM search_history WHERE user_id=$1`, userID)
	return err
}
