package port

import (
	"context"

	"github.com/google/uuid"
	"market-core/internal/core/domain"
	"market-core/internal/core/dto"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	List(ctx context.Context, filter dto.ProductFilter) ([]*domain.Product, int64, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
}
