package http

import (
	"log/slog"

	"market-core/internal/app/http/handler"
	"market-core/internal/app/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine *gin.Engine
	addr   string
}

type Deps struct {
	Products   *handler.ProductHandler
	Categories *handler.CategoryHandler
	Search     *handler.SearchHandler
	Health     *handler.HealthHandler
}

func NewServer(addr string, log *slog.Logger, deps Deps) *Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(middleware.Recover(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))

	r.GET("/health", deps.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		p := v1.Group("/products")
		p.POST("", deps.Products.Create)
		p.GET("", deps.Products.List)
		p.GET("/:id", deps.Products.Get)
		p.PUT("/:id", deps.Products.Update)
		p.DELETE("/:id", deps.Products.Delete)

		cat := v1.Group("/categories")
		cat.POST("", deps.Categories.Create)
		cat.GET("", deps.Categories.List)
		cat.GET("/:id", deps.Categories.Get)
		cat.DELETE("/:id", deps.Categories.Delete)

		s := v1.Group("/search")
		s.GET("", deps.Search.Search)
		s.GET("/autocomplete", deps.Search.Autocomplete)

		fav := v1.Group("/favorites")
		fav.GET("", deps.Search.ListFavorites)
		fav.POST("", deps.Search.AddFavorite)
		fav.DELETE("", deps.Search.RemoveFavorite)

		a := v1.Group("/analytics")
		a.GET("/popular-queries", deps.Search.PopularQueries)
		a.GET("/popular-products", deps.Search.PopularProducts)
	}

	return &Server{engine: r, addr: addr}
}

func (s *Server) ListenAndServe() error {
	return s.engine.Run(s.addr)
}
