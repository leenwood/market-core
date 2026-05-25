package http

import (
	"log/slog"
	"net/http"
	httppprof "net/http/pprof"

	internal "market-core/internal"
	"market-core/internal/app/http/handler"
	"market-core/internal/app/http/middleware"
	"market-core/internal/platform/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Deps struct {
	Products   *handler.ProductHandler
	Categories *handler.CategoryHandler
	Search     *handler.SearchHandler
	Health     *handler.HealthHandler
}

func NewServer(cfg internal.HTTPConfig, log *slog.Logger, m *metrics.Metrics, deps Deps) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(middleware.Recover(log))
	r.Use(middleware.Logger(log, m))
	r.Use(middleware.RequestID())
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MiB

	r.GET("/health", deps.Health.Health)
	r.GET("/ready", deps.Health.Ready)
	r.GET("/metrics", func(c *gin.Context) {
		promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if cfg.PprofEnabled {
		pp := r.Group("/debug/pprof")
		pp.GET("/", gin.WrapF(httppprof.Index))
		pp.GET("/cmdline", gin.WrapF(httppprof.Cmdline))
		pp.GET("/profile", gin.WrapF(httppprof.Profile))
		pp.GET("/symbol", gin.WrapF(httppprof.Symbol))
		pp.GET("/trace", gin.WrapF(httppprof.Trace))
		pp.GET("/:name", func(c *gin.Context) {
			httppprof.Handler(c.Param("name")).ServeHTTP(c.Writer, c.Request)
		})
	}

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

	// wrap gin engine with OTel handler, skipping /metrics and /health
	otelHandler := otelhttp.NewHandler(r, "server",
		otelhttp.WithFilter(func(req *http.Request) bool {
			p := req.URL.Path
			return p != "/metrics" && p != "/health"
		}),
	)

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      otelHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
