package handler

import (
	"errors"
	"net/http"

	"market-core/internal/core/domain"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard error envelope returned on all 4xx/5xx responses.
type ErrorResponse struct {
	Error string `json:"error" example:"not found"`
}

func writeJSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
	}
	c.JSON(status, ErrorResponse{Error: err.Error()})
}

func writeBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: msg})
}
