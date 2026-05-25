package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"market-core/internal/core/dto"
	productUC "market-core/internal/core/usecase/product"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	create *productUC.CreateUseCase
	update *productUC.UpdateUseCase
	delete *productUC.DeleteUseCase
	get    *productUC.GetUseCase
	list   *productUC.ListUseCase
}

func NewProductHandler(
	create *productUC.CreateUseCase,
	update *productUC.UpdateUseCase,
	del *productUC.DeleteUseCase,
	get *productUC.GetUseCase,
	list *productUC.ListUseCase,
) *ProductHandler {
	return &ProductHandler{create: create, update: update, delete: del, get: get, list: list}
}

// Create godoc
// @Summary      Create product
// @Description  Add a new product to the catalog
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateProductRequest  true  "Product data"
// @Success      201      {object}  dto.ProductResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse  "Category not found"
// @Failure      500      {object}  ErrorResponse
// @Router       /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}

	resp, err := h.create.Execute(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// Update godoc
// @Summary      Update product
// @Description  Update fields of an existing product (partial update)
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Product ID (UUID)"
// @Param        request  body      dto.UpdateProductRequest  true  "Fields to update"
// @Success      200      {object}  dto.ProductResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /products/{id} [put]
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeBadRequest(w, "invalid product id")
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}

	resp, err := h.update.Execute(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Delete product
// @Description  Soft-delete a product by ID
// @Tags         products
// @Produce      json
// @Param        id  path  string  true  "Product ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeBadRequest(w, "invalid product id")
		return
	}

	if err := h.delete.Execute(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Get godoc
// @Summary      Get product
// @Description  Return a single product by ID and increment its view counter
// @Tags         products
// @Produce      json
// @Param        id  path      string  true  "Product ID (UUID)"
// @Success      200  {object}  dto.ProductResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /products/{id} [get]
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeBadRequest(w, "invalid product id")
		return
	}

	resp, err := h.get.Execute(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// List godoc
// @Summary      List products
// @Description  Return a paginated list of products with optional filters
// @Tags         products
// @Produce      json
// @Param        category_id          query  string  false  "Filter by category UUID"
// @Param        include_subcategory  query  bool    false  "Include descendant subcategories"
// @Param        brand                query  string  false  "Filter by brand (partial match)"
// @Param        min_price            query  number  false  "Minimum price"
// @Param        max_price            query  number  false  "Maximum price"
// @Param        in_stock             query  bool    false  "Filter by stock availability"
// @Param        sort_by              query  string  false  "Sort field: price | created_at | popularity | rating"
// @Param        sort_dir             query  string  false  "Sort direction: asc | desc"
// @Param        page                 query  int     false  "Page number (default 1)"
// @Param        page_size            query  int     false  "Items per page (default 20, max 100)"
// @Success      200  {object}  dto.ProductListResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /products [get]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.ProductFilter{
		Page:     parseIntParam(r, "page", 1),
		PageSize: parseIntParam(r, "page_size", 20),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	if v := r.URL.Query().Get("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CategoryID = &id
			filter.IncludeSubcategory = r.URL.Query().Get("include_subcategory") == "true"
		}
	}
	if v := r.URL.Query().Get("brand"); v != "" {
		filter.Brand = &v
	}
	if v := r.URL.Query().Get("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = &f
		}
	}
	if v := r.URL.Query().Get("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = &f
		}
	}
	if v := r.URL.Query().Get("in_stock"); v != "" {
		b := v == "true"
		filter.InStock = &b
	}

	resp, err := h.list.Execute(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parseIntParam(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
