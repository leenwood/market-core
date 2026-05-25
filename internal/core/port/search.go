package port

import (
	"context"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"

	"github.com/google/uuid"
)

type SearchRepository interface {
	Search(ctx context.Context, req dto.SearchRequest) ([]*domain.Product, int64, error)
	Autocomplete(ctx context.Context, prefix string, limit int) ([]string, error)
	RecordQuery(ctx context.Context, q *domain.SearchQuery) error
	GetPopularQueries(ctx context.Context, limit int) ([]domain.PopularQuery, error)
	GetSearchHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.SearchHistory, error)
	ClearSearchHistory(ctx context.Context, userID uuid.UUID) error
}

type FavoritesRepository interface {
	Add(ctx context.Context, userID, productID uuid.UUID) error
	Remove(ctx context.Context, userID, productID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Product, int64, error)
}

type AnalyticsRepository interface {
	TrackClick(ctx context.Context, searchQueryID, productID uuid.UUID) error
	GetPopularProducts(ctx context.Context, limit int) ([]*domain.Product, error)
}
