package service

import (
	"context"
	"fmt"

	internal "market-core/internal"
	"market-core/internal/infra/storage/postgres"
	"market-core/internal/platform/metrics"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Infra struct {
	DB         *pgxpool.Pool
	Products   *postgres.ProductRepo
	Categories *postgres.CategoryRepo
	Search     *postgres.SearchRepo
	Analytics  *postgres.AnalyticsRepo
	Favorites  *postgres.FavoritesRepo
	Metrics    *metrics.Metrics
}

func initInfra(ctx context.Context, cfg *internal.Config) (*Infra, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Infra{
		DB:         db,
		Products:   postgres.NewProductRepo(db),
		Categories: postgres.NewCategoryRepo(db),
		Search:     postgres.NewSearchRepo(db),
		Analytics:  postgres.NewAnalyticsRepo(db),
		Favorites:  postgres.NewFavoritesRepo(db),
		Metrics:    metrics.New(),
	}, nil
}
