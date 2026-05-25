package service

import (
	"context"
	"fmt"
	"time"

	internal "market-core/internal"
	"market-core/internal/infra/storage/postgres"
	"market-core/internal/platform/metrics"
	"market-core/internal/platform/tracing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Infra struct {
	Cfg             *internal.Config
	DB              *pgxpool.Pool
	Products        *postgres.ProductRepo
	Categories      *postgres.CategoryRepo
	Search          *postgres.SearchRepo
	Analytics       *postgres.AnalyticsRepo
	Favorites       *postgres.FavoritesRepo
	Metrics         *metrics.Metrics
	shutdownTracing tracing.ShutdownFunc
}

func initInfra(ctx context.Context, cfg *internal.Config) (*Infra, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	shutdownTracing, err := tracing.Init(initCtx, tracing.Config{
		Enabled:     cfg.OTel.Enabled,
		Exporter:    cfg.OTel.Exporter,
		Endpoint:    cfg.OTel.Endpoint,
		ServiceName: cfg.OTel.ServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	db, err := pgxpool.New(initCtx, cfg.Postgres.DSN)
	if err != nil {
		_ = shutdownTracing(initCtx) //nolint:contextcheck
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := db.Ping(initCtx); err != nil {
		db.Close()
		_ = shutdownTracing(initCtx) //nolint:contextcheck
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Infra{
		Cfg:             cfg,
		DB:              db,
		Products:        postgres.NewProductRepo(db),
		Categories:      postgres.NewCategoryRepo(db),
		Search:          postgres.NewSearchRepo(db),
		Analytics:       postgres.NewAnalyticsRepo(db),
		Favorites:       postgres.NewFavoritesRepo(db),
		Metrics:         metrics.New(),
		shutdownTracing: shutdownTracing,
	}, nil
}

func (i *Infra) Shutdown(ctx context.Context) {
	i.DB.Close()
	_ = i.shutdownTracing(ctx)
}
