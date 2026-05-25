package handler

import (
	"encoding/json"
	"net/http"

	"market-core/internal/core/dto"
	categoryUC "market-core/internal/core/usecase/category"

	"github.com/go-chi/chi/v5"
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
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
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
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeBadRequest(w, "invalid category id")
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
// @Summary      List categories
// @Description  Return the full category tree (root nodes with nested children)
// @Tags         categories
// @Produce      json
// @Success      200  {array}   dto.CategoryResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.list.Execute(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
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
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeBadRequest(w, "invalid category id")
		return
	}

	if err := h.delete.Execute(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
