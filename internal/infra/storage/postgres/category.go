package postgres

import (
	"context"
	"errors"
	"time"

	"market-core/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO categories (id, name, slug, parent_id, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.ID, c.Name, c.Slug, c.ParentID, c.SortOrder, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	_, err := r.db.Exec(ctx, `
		UPDATE categories SET name=$2, slug=$3, sort_order=$4, updated_at=$5
		WHERE id=$1`,
		c.ID, c.Name, c.Slug, c.SortOrder, time.Now(),
	)
	return err
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE id=$1`, id)
	return scanCategory(row)
}

func (r *CategoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE slug=$1`, slug)
	return scanCategory(row)
}

func (r *CategoryRepo) GetTree(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, parent_id, sort_order, created_at, updated_at
		FROM categories ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	all := map[uuid.UUID]*domain.Category{}
	var roots []*domain.Category

	for rows.Next() {
		c, err := scanCategoryRows(rows)
		if err != nil {
			return nil, err
		}
		all[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, c := range all {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			if parent, ok := all[*c.ParentID]; ok {
				parent.Children = append(parent.Children, c)
			}
		}
	}
	return roots, nil
}

func (r *CategoryRepo) GetDescendantIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT id FROM categories WHERE id=$1
			UNION ALL
			SELECT c.id FROM categories c INNER JOIN descendants d ON c.parent_id=d.id
		)
		SELECT id FROM descendants`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var descID uuid.UUID
		if err := rows.Scan(&descID); err != nil {
			return nil, err
		}
		ids = append(ids, descID)
	}
	return ids, rows.Err()
}

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func scanCategoryRows(rows pgx.Rows) (*domain.Category, error) {
	var c domain.Category
	err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
