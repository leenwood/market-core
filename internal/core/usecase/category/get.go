package category

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"

	"github.com/google/uuid"
)

type GetUseCase struct {
	categories port.CategoryRepository
}

func NewGetUseCase(categories port.CategoryRepository) *GetUseCase {
	return &GetUseCase{categories: categories}
}

func (uc *GetUseCase) Execute(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	c, err := uc.categories.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return mapper.CategoryToResponse(c), nil
}
