package product

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"
)

type CreateUseCase struct {
	products   port.ProductRepository
	categories port.CategoryRepository
}

func NewCreateUseCase(products port.ProductRepository, categories port.CategoryRepository) *CreateUseCase {
	return &CreateUseCase{products: products, categories: categories}
}

func (uc *CreateUseCase) Execute(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	if _, err := uc.categories.GetByID(ctx, req.CategoryID); err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	p := mapper.CreateRequestToProduct(req)
	if err := uc.products.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return mapper.ProductToResponse(p), nil
}
