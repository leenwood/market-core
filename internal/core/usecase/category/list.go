package category

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"
)

type ListUseCase struct {
	categories port.CategoryRepository
}

func NewListUseCase(categories port.CategoryRepository) *ListUseCase {
	return &ListUseCase{categories: categories}
}

func (uc *ListUseCase) Execute(ctx context.Context) ([]*dto.CategoryResponse, error) {
	tree, err := uc.categories.GetTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("get category tree: %w", err)
	}
	return mapper.CategoriesToResponse(tree), nil
}
