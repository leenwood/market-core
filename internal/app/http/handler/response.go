package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"market-core/internal/core/domain"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data}) //nolint:errcheck
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Error: err.Error()}) //nolint:errcheck
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(envelope{Error: msg}) //nolint:errcheck
}
