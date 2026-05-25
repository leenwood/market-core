package http

import (
	"net/http"

	"log/slog"
	"market-core/internal/app/http/handler"
	"market-core/internal/app/http/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	router http.Handler
	addr   string
}

type Deps struct {
	Products   *handler.ProductHandler
	Categories *handler.CategoryHandler
	Search     *handler.SearchHandler
	Health     *handler.HealthHandler
}

func NewServer(addr string, log *slog.Logger, deps Deps) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Recover(log))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(chiMiddleware.Compress(5))

	r.Get("/health", deps.Health.Health)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Post("/", deps.Products.Create)
			r.Get("/", deps.Products.List)
			r.Get("/{id}", deps.Products.Get)
			r.Put("/{id}", deps.Products.Update)
			r.Delete("/{id}", deps.Products.Delete)
		})

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", deps.Categories.Create)
			r.Get("/", deps.Categories.List)
			r.Get("/{id}", deps.Categories.Get)
			r.Delete("/{id}", deps.Categories.Delete)
		})

		r.Route("/search", func(r chi.Router) {
			r.Get("/", deps.Search.Search)
			r.Get("/autocomplete", deps.Search.Autocomplete)
		})

		r.Route("/favorites", func(r chi.Router) {
			r.Get("/", deps.Search.ListFavorites)
			r.Post("/", deps.Search.AddFavorite)
			r.Delete("/", deps.Search.RemoveFavorite)
		})

		r.Route("/analytics", func(r chi.Router) {
			r.Get("/popular-queries", deps.Search.PopularQueries)
			r.Get("/popular-products", deps.Search.PopularProducts)
		})
	})

	return &Server{router: r, addr: addr}
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.router)
}
