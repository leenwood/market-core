package category

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"market-core/internal/core/port"
)

type DeleteUseCase struct {
	categories port.CategoryRepository
}

func NewDeleteUseCase(categories port.CategoryRepository) *DeleteUseCase {
	return &DeleteUseCase{categories: categories}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.categories.GetByID(ctx, id); err != nil {
		return fmt.Errorf("get category: %w", err)
	}
	return uc.categories.Delete(ctx, id)
}
