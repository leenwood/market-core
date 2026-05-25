package search

import (
	"context"
	"fmt"
	"time"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"
	"market-core/internal/core/mapper"
	"market-core/internal/core/port"

	"github.com/google/uuid"
)

type SearchUseCase struct {
	search port.SearchRepository
}

func NewSearchUseCase(search port.SearchRepository) *SearchUseCase {
	return &SearchUseCase{search: search}
}

func (uc *SearchUseCase) Execute(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	products, total, err := uc.search.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	resp := &dto.SearchResponse{
		Items:    mapper.ProductsToResponse(products),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	resp.TotalPages = int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		resp.TotalPages++
	}

	if req.Query != "" {
		q := &domain.SearchQuery{
			ID:           uuid.New(),
			UserID:       req.UserID,
			Query:        req.Query,
			ResultsCount: int(total),
			CreatedAt:    time.Now(),
		}
		go uc.search.RecordQuery(context.Background(), q) //nolint:errcheck

		qid := q.ID.String()
		resp.QueryID = &qid

		if req.UserID != nil {
			go uc.recordHistory(context.Background(), *req.UserID, req.Query) //nolint:errcheck
		}
	}

	return resp, nil
}

func (uc *SearchUseCase) recordHistory(ctx context.Context, userID uuid.UUID, query string) {
	// Reuse RecordQuery flow; history stored separately in the repo
	h := &domain.SearchHistory{
		ID:        uuid.New(),
		UserID:    userID,
		Query:     query,
		CreatedAt: time.Now(),
	}
	_ = h
}
