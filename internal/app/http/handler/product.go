package handler

import (
	"net/http"
	"strconv"

	"market-core/internal/core/dto"
	productUC "market-core/internal/core/usecase/product"

	"github.com/gin-gonic/gin"
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
// @Router       /products [post].
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid request body")
		return
	}

	resp, err := h.create.Execute(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, resp)
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
// @Router       /products/{id} [put].
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		writeBadRequest(c, "invalid product id")
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, "invalid request body")
		return
	}

	resp, err := h.update.Execute(c.Request.Context(), id, req)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
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
// @Router       /products/{id} [delete].
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		writeBadRequest(c, "invalid product id")
		return
	}

	if err := h.delete.Execute(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
// @Router       /products/{id} [get].
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		writeBadRequest(c, "invalid product id")
		return
	}

	resp, err := h.get.Execute(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
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
// @Router       /products [get].
func (h *ProductHandler) List(c *gin.Context) {
	filter := dto.ProductFilter{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "page_size", 20),
		SortBy:   c.Query("sort_by"),
		SortDir:  c.Query("sort_dir"),
	}

	if v := c.Query("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CategoryID = &id
			filter.IncludeSubcategory = c.Query("include_subcategory") == "true"
		}
	}
	if v := c.Query("brand"); v != "" {
		filter.Brand = &v
	}
	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = &f
		}
	}
	if v := c.Query("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = &f
		}
	}
	if v := c.Query("in_stock"); v != "" {
		b := v == "true"
		filter.InStock = &b
	}

	resp, err := h.list.Execute(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parseIntParam(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
