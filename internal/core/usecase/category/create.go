package category

import (
	"context"
	"fmt"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"
)

type CreateUseCase struct {
	categories port.CategoryRepository
}

func NewCreateUseCase(categories port.CategoryRepository) *CreateUseCase {
	return &CreateUseCase{categories: categories}
}

func (uc *CreateUseCase) Execute(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	existing, err := uc.categories.GetBySlug(ctx, req.Slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("slug already taken: %w", domain.ErrAlreadyExists)
	}

	if req.ParentID != nil {
		if _, err := uc.categories.GetByID(ctx, *req.ParentID); err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
	}

	c := mapper.CreateRequestToCategory(req)
	if err := uc.categories.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return mapper.CategoryToResponse(c), nil
}
