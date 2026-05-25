package handler

import (
	"net/http"
	"strconv"

	"market-core/internal/core/dto"
	analyticsUC "market-core/internal/core/usecase/analytics"
	favoritesUC "market-core/internal/core/usecase/favorites"
	searchUC "market-core/internal/core/usecase/search"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SearchHandler struct {
	search       *searchUC.SearchUseCase
	autocomplete *searchUC.AutocompleteUseCase
	analytics    *analyticsUC.TrackUseCase
	favorites    *favoritesUC.UseCase
}

func NewSearchHandler(
	search *searchUC.SearchUseCase,
	autocomplete *searchUC.AutocompleteUseCase,
	analytics *analyticsUC.TrackUseCase,
	favorites *favoritesUC.UseCase,
) *SearchHandler {
	return &SearchHandler{
		search:       search,
		autocomplete: autocomplete,
		analytics:    analytics,
		favorites:    favorites,
	}
}

// Search godoc
// @Summary      Search products
// @Description  Full-text search combined with SQL filters. Supports fuzzy matching via trigram similarity.
// @Description  Results are ranked by: ts_rank_cd × 0.7 + word_similarity × 0.3.
// @Description  Pass X-User-ID header to record search history.
// @Tags         search
// @Produce      json
// @Param        q            query   string  false  "Search query"
// @Param        category_id  query   string  false  "Filter by category UUID (includes subcategories)"
// @Param        brand        query   string  false  "Filter by brand (partial match)"
// @Param        min_price    query   number  false  "Minimum price"
// @Param        max_price    query   number  false  "Maximum price"
// @Param        in_stock     query   bool    false  "Filter by stock availability"
// @Param        sort_by      query   string  false  "relevance | price | created_at | popularity"
// @Param        sort_dir     query   string  false  "asc | desc"
// @Param        page         query   int     false  "Page number (default 1)"
// @Param        page_size    query   int     false  "Items per page (default 20, max 100)"
// @Param        X-User-ID    header  string  false  "User ID for search history recording"
// @Success      200  {object}  dto.SearchResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	req := dto.SearchRequest{
		Query:    c.Query("q"),
		SortBy:   c.Query("sort_by"),
		SortDir:  c.Query("sort_dir"),
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "page_size", 20),
	}

	if v := c.Query("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.CategoryID = &id
		}
	}
	if v := c.Query("brand"); v != "" {
		req.Brand = &v
	}
	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MinPrice = &f
		}
	}
	if v := c.Query("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MaxPrice = &f
		}
	}
	if v := c.Query("in_stock"); v != "" {
		b := v == "true"
		req.InStock = &b
	}
	if v := c.GetHeader("X-User-ID"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.UserID = &id
		}
	}

	resp, err := h.search.Execute(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
}

// Autocomplete godoc
// @Summary      Search autocomplete
// @Description  Return up to `limit` suggestions matching the given prefix from product names and past queries
// @Tags         search
// @Produce      json
// @Param        q      query  string  true   "Prefix string"
// @Param        limit  query  int     false  "Max suggestions (default 10, max 20)"
// @Success      200  {object}  dto.AutocompleteResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /search/autocomplete [get]
func (h *SearchHandler) Autocomplete(c *gin.Context) {
	prefix := c.Query("q")
	if prefix == "" {
		writeBadRequest(c, "q parameter is required")
		return
	}

	req := dto.AutocompleteRequest{
		Prefix: prefix,
		Limit:  parseIntParam(c, "limit", 10),
	}

	resp, err := h.autocomplete.Execute(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
}

// PopularQueries godoc
// @Summary      Popular search queries
// @Description  Return the most frequent search queries over the last 30 days
// @Tags         analytics
// @Produce      json
// @Param        limit  query  int  false  "Number of results (default 10, max 100)"
// @Success      200  {array}   dto.PopularQueryResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /analytics/popular-queries [get]
func (h *SearchHandler) PopularQueries(c *gin.Context) {
	limit := parseIntParam(c, "limit", 10)
	queries, err := h.analytics.GetPopularQueries(c.Request.Context(), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, queries)
}

// PopularProducts godoc
// @Summary      Popular products
// @Description  Return the most viewed products sorted by view count descending
// @Tags         analytics
// @Produce      json
// @Param        limit  query  int  false  "Number of results (default 10, max 100)"
// @Success      200  {array}   dto.ProductResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /analytics/popular-products [get]
func (h *SearchHandler) PopularProducts(c *gin.Context) {
	limit := parseIntParam(c, "limit", 10)
	products, err := h.analytics.GetPopularProducts(c.Request.Context(), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, products)
}

// AddFavorite godoc
// @Summary      Add to favorites
// @Description  Add a product to the authenticated user's favorites list
// @Tags         favorites
// @Produce      json
// @Param        X-User-ID   header  string  true  "User ID (UUID)"
// @Param        product_id  query   string  true  "Product ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse  "Product not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /favorites [post]
func (h *SearchHandler) AddFavorite(c *gin.Context) {
	userID, err := requireUserID(c)
	if err != nil {
		writeBadRequest(c, "X-User-ID header required")
		return
	}
	productID, err := parseUUID(c.Query("product_id"))
	if err != nil {
		writeBadRequest(c, "invalid product_id")
		return
	}
	if err := h.favorites.Add(c.Request.Context(), userID, productID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveFavorite godoc
// @Summary      Remove from favorites
// @Description  Remove a product from the authenticated user's favorites list
// @Tags         favorites
// @Produce      json
// @Param        X-User-ID   header  string  true  "User ID (UUID)"
// @Param        product_id  query   string  true  "Product ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /favorites [delete]
func (h *SearchHandler) RemoveFavorite(c *gin.Context) {
	userID, err := requireUserID(c)
	if err != nil {
		writeBadRequest(c, "X-User-ID header required")
		return
	}
	productID, err := parseUUID(c.Query("product_id"))
	if err != nil {
		writeBadRequest(c, "invalid product_id")
		return
	}
	if err := h.favorites.Remove(c.Request.Context(), userID, productID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListFavorites godoc
// @Summary      List favorites
// @Description  Return a paginated list of the user's favorite products
// @Tags         favorites
// @Produce      json
// @Param        X-User-ID  header  string  true   "User ID (UUID)"
// @Param        page       query   int     false  "Page number (default 1)"
// @Param        page_size  query   int     false  "Items per page (default 20, max 100)"
// @Success      200  {object}  dto.ProductListResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /favorites [get]
func (h *SearchHandler) ListFavorites(c *gin.Context) {
	userID, err := requireUserID(c)
	if err != nil {
		writeBadRequest(c, "X-User-ID header required")
		return
	}
	page := parseIntParam(c, "page", 1)
	pageSize := parseIntParam(c, "page_size", 20)

	resp, err := h.favorites.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
}

func requireUserID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetHeader("X-User-ID"))
}
