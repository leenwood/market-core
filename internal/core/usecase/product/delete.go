package product

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"market-core/internal/core/port"
)

type DeleteUseCase struct {
	products port.ProductRepository
}

func NewDeleteUseCase(products port.ProductRepository) *DeleteUseCase {
	return &DeleteUseCase{products: products}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.products.GetByID(ctx, id); err != nil {
		return fmt.Errorf("get product: %w", err)
	}
	return uc.products.Delete(ctx, id)
}
