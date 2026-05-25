package favorites

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"
)

type UseCase struct {
	favorites port.FavoritesRepository
	products  port.ProductRepository
}

func NewUseCase(favorites port.FavoritesRepository, products port.ProductRepository) *UseCase {
	return &UseCase{favorites: favorites, products: products}
}

func (uc *UseCase) Add(ctx context.Context, userID, productID uuid.UUID) error {
	if _, err := uc.products.GetByID(ctx, productID); err != nil {
		return fmt.Errorf("product not found: %w", err)
	}
	return uc.favorites.Add(ctx, userID, productID)
}

func (uc *UseCase) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	return uc.favorites.Remove(ctx, userID, productID)
}

func (uc *UseCase) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.ProductListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := uc.favorites.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &dto.ProductListResponse{
		Items:      mapper.ProductsToResponse(products),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
