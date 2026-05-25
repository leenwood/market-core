package dto

import "github.com/google/uuid"

type SearchRequest struct {
	Query      string         `json:"query"`
	CategoryID *uuid.UUID     `json:"category_id"`
	Brand      *string        `json:"brand"`
	MinPrice   *float64       `json:"min_price"`
	MaxPrice   *float64       `json:"max_price"`
	InStock    *bool          `json:"in_stock"`
	Attributes map[string]any `json:"attributes"`
	SortBy     string         `json:"sort_by"` // relevance, price, created_at, popularity
	SortDir    string         `json:"sort_dir"` // asc, desc
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	UserID     *uuid.UUID     `json:"-"`
}

type SearchResponse struct {
	Items      []*ProductResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	QueryID    *string            `json:"query_id,omitempty"`
}

type AutocompleteRequest struct {
	Prefix string `json:"prefix" validate:"required,min=1,max=100"`
	Limit  int    `json:"limit"`
}

type AutocompleteResponse struct {
	Suggestions []string `json:"suggestions"`
}

type PopularQueryResponse struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

type SearchHistoryItem struct {
	ID        string `json:"id"`
	Query     string `json:"query"`
	CreatedAt string `json:"created_at"`
}

type TrackClickRequest struct {
	SearchQueryID string `json:"search_query_id" validate:"required"`
	ProductID     string `json:"product_id" validate:"required"`
}
