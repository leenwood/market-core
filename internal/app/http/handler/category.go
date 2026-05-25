package handler

import (
	"net/http"

	"market-core/internal/core/dto"
	categoryUC "market-core/internal/core/usecase/category"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	create *categoryUC.CreateUseCase
	get    *categoryUC.GetUseCase
	list   *categoryUC.ListUseCase
	delete *categoryUC.DeleteUseCase
}

func NewCategoryHandler(
	create *categoryUC.CreateUseCase,
	get *categoryUC.GetUseCase,
	list *categoryUC.ListUseCase,
	del *categoryUC.DeleteUseCase,
) *CategoryHandler {
	return &CategoryHandler{create: create, get: get, list: list, delete: del}
}

// Create godoc
// @Summary      Create category
// @Description  Create a new category; set parent_id to nest it in the hierarchy
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateCategoryRequest  true  "Category data"
// @Success      201      {object}  dto.CategoryResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse  "Parent category not found"
// @Failure      409      {object}  ErrorResponse  "Slug already taken"
// @Failure      500      {object}  ErrorResponse
// @Router       /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
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

// Get godoc
// @Summary      Get category
// @Description  Return a single category by ID
// @Tags         categories
// @Produce      json
// @Param        id  path      string  true  "Category ID (UUID)"
// @Success      200  {object}  dto.CategoryResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		writeBadRequest(c, "invalid category id")
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
// @Summary      List categories
// @Description  Return the full category tree (root nodes with nested children)
// @Tags         categories
// @Produce      json
// @Success      200  {array}   dto.CategoryResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	resp, err := h.list.Execute(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Delete category
// @Description  Delete a category by ID. Child categories will have their parent_id set to NULL.
// @Tags         categories
// @Produce      json
// @Param        id  path  string  true  "Category ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		writeBadRequest(c, "invalid category id")
		return
	}

	if err := h.delete.Execute(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
