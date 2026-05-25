package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	internal "market-core/internal"
	apphttp "market-core/internal/app/http"
	"market-core/internal/app/http/handler"
	analyticsUC "market-core/internal/core/usecase/analytics"
	categoryUC "market-core/internal/core/usecase/category"
	favoritesUC "market-core/internal/core/usecase/favorites"
	productUC "market-core/internal/core/usecase/product"
	searchUC "market-core/internal/core/usecase/search"
	"market-core/internal/platform/logger"
)

func RunServer(ctx context.Context) error {
	cfg, err := internal.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	infra, err := initInfra(ctx, cfg)
	if err != nil {
		return err
	}

	// product use cases
	createProduct := productUC.NewCreateUseCase(infra.Products, infra.Categories)
	updateProduct := productUC.NewUpdateUseCase(infra.Products, infra.Categories)
	deleteProduct := productUC.NewDeleteUseCase(infra.Products)
	getProduct := productUC.NewGetUseCase(infra.Products)
	listProducts := productUC.NewListUseCase(infra.Products)

	// category use cases
	createCategory := categoryUC.NewCreateUseCase(infra.Categories)
	getCategory := categoryUC.NewGetUseCase(infra.Categories)
	listCategories := categoryUC.NewListUseCase(infra.Categories)
	deleteCategory := categoryUC.NewDeleteUseCase(infra.Categories)

	// search / analytics / favorites use cases
	search := searchUC.NewSearchUseCase(infra.Search)
	autocomplete := searchUC.NewAutocompleteUseCase(infra.Search)
	track := analyticsUC.NewTrackUseCase(infra.Analytics, infra.Search, infra.Products)
	favorites := favoritesUC.NewUseCase(infra.Favorites, infra.Products)

	productHandler := handler.NewProductHandler(createProduct, updateProduct, deleteProduct, getProduct, listProducts)
	categoryHandler := handler.NewCategoryHandler(createCategory, getCategory, listCategories, deleteCategory)
	searchHandler := handler.NewSearchHandler(search, autocomplete, track, favorites)
	healthHandler := handler.NewHealthHandler(infra.DB)

	srv := apphttp.NewServer(cfg.HTTP, log, infra.Metrics, apphttp.Deps{
		Products:   productHandler,
		Categories: categoryHandler,
		Search:     searchHandler,
		Health:     healthHandler,
	})

	log.Info("server started", "addr", cfg.HTTP.Addr)

	srvErr := make(chan error, 1)
	go func() { srvErr <- srv.ListenAndServe() }()

	select {
	case err := <-srvErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown error", "error", err)
	}
	infra.Shutdown(shutdownCtx)

	return nil
}
