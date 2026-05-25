package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	attrs, err := json.Marshal(p.Attributes)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO products (id, name, description, category_id, brand, price, in_stock, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, p.Name, p.Description, p.CategoryID, p.Brand, p.Price, p.InStock, attrs, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	attrs, err := json.Marshal(p.Attributes)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE products
		SET name=$2, description=$3, category_id=$4, brand=$5, price=$6, in_stock=$7, attributes=$8, updated_at=$9
		WHERE id=$1 AND deleted_at IS NULL`,
		p.ID, p.Name, p.Description, p.CategoryID, p.Brand, p.Price, p.InStock, attrs, p.UpdatedAt,
	)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE products SET deleted_at=$2 WHERE id=$1 AND deleted_at IS NULL`,
		id, time.Now(),
	)
	return err
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, description, category_id, brand, price, rating, rating_count, in_stock,
		       attributes, view_count, created_at, updated_at, deleted_at
		FROM products WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanProduct(row)
}

func (r *ProductRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE products SET view_count=view_count+1 WHERE id=$1`, id)
	return err
}

func (r *ProductRepo) List(ctx context.Context, f dto.ProductFilter) ([]*domain.Product, int64, error) {
	conditions := []string{"p.deleted_at IS NULL"}
	args := []any{}
	n := 1

	if f.CategoryID != nil {
		if f.IncludeSubcategory {
			conditions = append(conditions, fmt.Sprintf(`p.category_id IN (
				WITH RECURSIVE cats AS (
					SELECT id FROM categories WHERE id=$%d
					UNION ALL
					SELECT c.id FROM categories c INNER JOIN cats ON c.parent_id=cats.id
				) SELECT id FROM cats)`, n))
		} else {
			conditions = append(conditions, fmt.Sprintf("p.category_id=$%d", n))
		}
		args = append(args, *f.CategoryID)
		n++
	}
	if f.Brand != nil {
		conditions = append(conditions, fmt.Sprintf("p.brand ILIKE $%d", n))
		args = append(args, "%"+*f.Brand+"%")
		n++
	}
	if f.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price>=$%d", n))
		args = append(args, *f.MinPrice)
		n++
	}
	if f.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price<=$%d", n))
		args = append(args, *f.MaxPrice)
		n++
	}
	if f.InStock != nil {
		conditions = append(conditions, fmt.Sprintf("p.in_stock=$%d", n))
		args = append(args, *f.InStock)
		n++
	}
	if len(f.Attributes) > 0 {
		for k, v := range f.Attributes {
			conditions = append(conditions, fmt.Sprintf("p.attributes->>'%s'=$%d", sanitizeKey(k), n))
			args = append(args, fmt.Sprintf("%v", v))
			n++
		}
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	orderBy := buildProductOrderBy(f.SortBy, f.SortDir)
	offset := (f.Page - 1) * f.PageSize

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products p %s`, where)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.PageSize, offset)
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.category_id, p.brand, p.price, p.rating, p.rating_count,
		       p.in_stock, p.attributes, p.view_count, p.created_at, p.updated_at, p.deleted_at
		FROM products p %s %s LIMIT $%d OFFSET $%d`, where, orderBy, n, n+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p, err := scanProductRows(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func buildProductOrderBy(sortBy, sortDir string) string {
	dir := "ASC"
	if strings.ToLower(sortDir) == "desc" {
		dir = "DESC"
	}
	switch sortBy {
	case "price":
		return fmt.Sprintf("ORDER BY p.price %s", dir)
	case "created_at":
		return fmt.Sprintf("ORDER BY p.created_at %s", dir)
	case "popularity":
		return "ORDER BY p.view_count DESC"
	case "rating":
		return fmt.Sprintf("ORDER BY p.rating %s", dir)
	default:
		return "ORDER BY p.created_at DESC"
	}
}

func sanitizeKey(k string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, k)
}

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var attrsRaw []byte
	err := row.Scan(
		&p.ID, &p.Name, &p.Description, &p.CategoryID, &p.Brand, &p.Price,
		&p.Rating, &p.RatingCount, &p.InStock, &attrsRaw, &p.ViewCount,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
		p.Attributes = map[string]any{}
	}
	return &p, nil
}

func scanProductRows(rows pgx.Rows) (*domain.Product, error) {
	var p domain.Product
	var attrsRaw []byte
	err := rows.Scan(
		&p.ID, &p.Name, &p.Description, &p.CategoryID, &p.Brand, &p.Price,
		&p.Rating, &p.RatingCount, &p.InStock, &attrsRaw, &p.ViewCount,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(attrsRaw, &p.Attributes); err != nil {
		p.Attributes = map[string]any{}
	}
	return &p, nil
}
