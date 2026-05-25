package product

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"

	"github.com/google/uuid"
)

type GetUseCase struct {
	products port.ProductRepository
}

func NewGetUseCase(products port.ProductRepository) *GetUseCase {
	return &GetUseCase{products: products}
}

func (uc *GetUseCase) Execute(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	p, err := uc.products.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	go uc.products.IncrementViewCount(context.Background(), id) //nolint:errcheck

	return mapper.ProductToResponse(p), nil
}
