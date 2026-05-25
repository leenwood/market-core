package handler

import (
	"net/http"
	"strconv"

	"market-core/internal/core/dto"
	analyticsUC "market-core/internal/core/usecase/analytics"
	favoritesUC "market-core/internal/core/usecase/favorites"
	searchUC "market-core/internal/core/usecase/search"

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
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	req := dto.SearchRequest{
		Query:    r.URL.Query().Get("q"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
		Page:     parseIntParam(r, "page", 1),
		PageSize: parseIntParam(r, "page_size", 20),
	}

	if v := r.URL.Query().Get("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.CategoryID = &id
		}
	}
	if v := r.URL.Query().Get("brand"); v != "" {
		req.Brand = &v
	}
	if v := r.URL.Query().Get("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MinPrice = &f
		}
	}
	if v := r.URL.Query().Get("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MaxPrice = &f
		}
	}
	if v := r.URL.Query().Get("in_stock"); v != "" {
		b := v == "true"
		req.InStock = &b
	}
	if v := r.Header.Get("X-User-ID"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.UserID = &id
		}
	}

	resp, err := h.search.Execute(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
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
func (h *SearchHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	if prefix == "" {
		writeBadRequest(w, "q parameter is required")
		return
	}

	req := dto.AutocompleteRequest{
		Prefix: prefix,
		Limit:  parseIntParam(r, "limit", 10),
	}

	resp, err := h.autocomplete.Execute(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
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
func (h *SearchHandler) PopularQueries(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 10)
	queries, err := h.analytics.GetPopularQueries(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, queries)
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
func (h *SearchHandler) PopularProducts(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 10)
	products, err := h.analytics.GetPopularProducts(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
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
func (h *SearchHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := requireUserID(r)
	if err != nil {
		writeBadRequest(w, "X-User-ID header required")
		return
	}
	productID, err := parseUUID(r.URL.Query().Get("product_id"))
	if err != nil {
		writeBadRequest(w, "invalid product_id")
		return
	}
	if err := h.favorites.Add(r.Context(), userID, productID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
func (h *SearchHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := requireUserID(r)
	if err != nil {
		writeBadRequest(w, "X-User-ID header required")
		return
	}
	productID, err := parseUUID(r.URL.Query().Get("product_id"))
	if err != nil {
		writeBadRequest(w, "invalid product_id")
		return
	}
	if err := h.favorites.Remove(r.Context(), userID, productID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
func (h *SearchHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID, err := requireUserID(r)
	if err != nil {
		writeBadRequest(w, "X-User-ID header required")
		return
	}
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "page_size", 20)

	resp, err := h.favorites.List(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func requireUserID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.Header.Get("X-User-ID"))
}
