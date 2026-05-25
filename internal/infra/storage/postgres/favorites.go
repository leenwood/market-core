package postgres

import (
	"context"
	"encoding/json"
	"time"

	"market-core/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FavoritesRepo struct {
	db *pgxpool.Pool
}

func NewFavoritesRepo(db *pgxpool.Pool) *FavoritesRepo {
	return &FavoritesRepo{db: db}
}

func (r *FavoritesRepo) Add(ctx context.Context, userID, productID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO favorites (user_id, product_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		userID, productID, time.Now(),
	)
	return err
}

func (r *FavoritesRepo) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM favorites WHERE user_id=$1 AND product_id=$2`, userID, productID)
	return err
}

func (r *FavoritesRepo) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Product, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM favorites WHERE user_id=$1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, p.description, p.category_id, p.brand, p.price, p.rating, p.rating_count,
		       p.in_stock, p.attributes, p.view_count, p.created_at, p.updated_at, p.deleted_at
		FROM products p
		INNER JOIN favorites f ON f.product_id=p.id
		WHERE f.user_id=$1 AND p.deleted_at IS NULL
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
			p.Attributes = map[string]any{}
		}
		products = append(products, &p)
	}
	return products, total, rows.Err()
}
