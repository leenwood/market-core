package port

import (
	"context"

	"market-core/internal/core/domain"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *domain.Category) error
	Update(ctx context.Context, c *domain.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetTree(ctx context.Context) ([]*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	GetDescendantIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
}
