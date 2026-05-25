package service

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	infra, err := initInfra(ctx, cfg)
	if err != nil {
		return err
	}
	defer infra.DB.Close()

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

	// handlers
	productHandler := handler.NewProductHandler(createProduct, updateProduct, deleteProduct, getProduct, listProducts)
	categoryHandler := handler.NewCategoryHandler(createCategory, getCategory, listCategories, deleteCategory)
	searchHandler := handler.NewSearchHandler(search, autocomplete, track, favorites)
	healthHandler := handler.NewHealthHandler()

	srv := apphttp.NewServer(cfg.HTTPAddr, log, apphttp.Deps{
		Products:   productHandler,
		Categories: categoryHandler,
		Search:     searchHandler,
		Health:     healthHandler,
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	log.Info("server started", "addr", cfg.HTTPAddr)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		log.Info("shutting down")
		return nil
	}
}
