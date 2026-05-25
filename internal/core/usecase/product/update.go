package product

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"

	"github.com/google/uuid"
)

type UpdateUseCase struct {
	products   port.ProductRepository
	categories port.CategoryRepository
}

func NewUpdateUseCase(products port.ProductRepository, categories port.CategoryRepository) *UpdateUseCase {
	return &UpdateUseCase{products: products, categories: categories}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	p, err := uc.products.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	if req.CategoryID != nil {
		if _, err := uc.categories.GetByID(ctx, *req.CategoryID); err != nil {
			return nil, fmt.Errorf("category not found: %w", err)
		}
	}

	mapper.ApplyUpdateRequest(p, req)

	if err := uc.products.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return mapper.ProductToResponse(p), nil
}
