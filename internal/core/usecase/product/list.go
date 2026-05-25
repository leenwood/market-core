package product

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"
)

type ListUseCase struct {
	products port.ProductRepository
}

func NewListUseCase(products port.ProductRepository) *ListUseCase {
	return &ListUseCase{products: products}
}

func (uc *ListUseCase) Execute(ctx context.Context, filter dto.ProductFilter) (*dto.ProductListResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	products, total, err := uc.products.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize != 0 {
		totalPages++
	}

	return &dto.ProductListResponse{
		Items:      mapper.ProductsToResponse(products),
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}
