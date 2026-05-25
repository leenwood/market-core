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

func (h *SearchHandler) PopularQueries(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 10)
	queries, err := h.analytics.GetPopularQueries(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, queries)
}

func (h *SearchHandler) PopularProducts(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 10)
	products, err := h.analytics.GetPopularProducts(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

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
