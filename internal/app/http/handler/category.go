package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"market-core/internal/core/dto"
	categoryUC "market-core/internal/core/usecase/category"
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

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.list.Execute(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

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
