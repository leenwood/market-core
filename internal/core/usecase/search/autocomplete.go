package search

import (
	"context"
	"fmt"

	"market-core/internal/core/dto"
	"market-core/internal/core/port"
)

type AutocompleteUseCase struct {
	search port.SearchRepository
}

func NewAutocompleteUseCase(search port.SearchRepository) *AutocompleteUseCase {
	return &AutocompleteUseCase{search: search}
}

func (uc *AutocompleteUseCase) Execute(ctx context.Context, req dto.AutocompleteRequest) (*dto.AutocompleteResponse, error) {
	limit := req.Limit
	if limit < 1 || limit > 20 {
		limit = 10
	}

	suggestions, err := uc.search.Autocomplete(ctx, req.Prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("autocomplete: %w", err)
	}

	return &dto.AutocompleteResponse{Suggestions: suggestions}, nil
}
