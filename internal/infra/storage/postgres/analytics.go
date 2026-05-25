package postgres

import (
	"context"
	"encoding/json"

	"market-core/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepo struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepo(db *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) TrackClick(ctx context.Context, searchQueryID, productID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO search_clicks (id, search_query_id, product_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())`,
		searchQueryID, productID,
	)
	return err
}

func (r *AnalyticsRepo) GetPopularProducts(ctx context.Context, limit int) ([]*domain.Product, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, p.description, p.category_id, p.brand, p.price, p.rating, p.rating_count,
		       p.in_stock, p.attributes, p.view_count, p.created_at, p.updated_at, p.deleted_at
		FROM products p
		WHERE p.deleted_at IS NULL
		ORDER BY p.view_count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		var attrsRaw []byte
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.CategoryID, &p.Brand, &p.Price,
			&p.Rating, &p.RatingCount, &p.InStock, &attrsRaw, &p.ViewCount,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
			p.Attributes = map[string]any{}
		}
		products = append(products, &p)
	}
	return products, rows.Err()
}
