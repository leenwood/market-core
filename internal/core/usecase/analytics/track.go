package analytics

import (
	"context"
	"fmt"

	"market-core/internal/core/port"

	"github.com/google/uuid"
)

type TrackUseCase struct {
	analytics port.AnalyticsRepository
	search    port.SearchRepository
	products  port.ProductRepository
}

func NewTrackUseCase(analytics port.AnalyticsRepository, search port.SearchRepository, products port.ProductRepository) *TrackUseCase {
	return &TrackUseCase{analytics: analytics, search: search, products: products}
}

func (uc *TrackUseCase) TrackClick(ctx context.Context, searchQueryID, productID uuid.UUID) error {
	if _, err := uc.products.GetByID(ctx, productID); err != nil {
		return fmt.Errorf("product not found: %w", err)
	}
	return uc.analytics.TrackClick(ctx, searchQueryID, productID)
}

func (uc *TrackUseCase) GetPopularQueries(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	queries, err := uc.search.GetPopularQueries(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("get popular queries: %w", err)
	}

	result := make([]map[string]any, len(queries))
	for i, q := range queries {
		result[i] = map[string]any{
			"query": q.Query,
			"count": q.Count,
		}
	}
	return result, nil
}

func (uc *TrackUseCase) GetPopularProducts(ctx context.Context, limit int) (interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return uc.analytics.GetPopularProducts(ctx, limit)
}
