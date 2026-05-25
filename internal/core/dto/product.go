package dto

import "github.com/google/uuid"

type CreateProductRequest struct {
	Name        string         `json:"name" validate:"required,min=1,max=500"`
	Description string         `json:"description"`
	CategoryID  uuid.UUID      `json:"category_id" validate:"required"`
	Brand       string         `json:"brand"`
	Price       float64        `json:"price" validate:"gte=0"`
	InStock     bool           `json:"in_stock"`
	Attributes  map[string]any `json:"attributes"`
}

type UpdateProductRequest struct {
	Name        *string        `json:"name" validate:"omitempty,min=1,max=500"`
	Description *string        `json:"description"`
	CategoryID  *uuid.UUID     `json:"category_id"`
	Brand       *string        `json:"brand"`
	Price       *float64       `json:"price" validate:"omitempty,gte=0"`
	InStock     *bool          `json:"in_stock"`
	Attributes  map[string]any `json:"attributes"`
}

type ProductFilter struct {
	CategoryID         *uuid.UUID
	IncludeSubcategory bool
	Brand              *string
	MinPrice           *float64
	MaxPrice           *float64
	InStock            *bool
	Attributes         map[string]any
	Page               int
	PageSize           int
	SortBy             string
	SortDir            string
}

type ProductResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CategoryID  string         `json:"category_id"`
	Brand       string         `json:"brand"`
	Price       float64        `json:"price"`
	Rating      float64        `json:"rating"`
	RatingCount int            `json:"rating_count"`
	InStock     bool           `json:"in_stock"`
	Attributes  map[string]any `json:"attributes"`
	ViewCount   int64          `json:"view_count"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type ProductListResponse struct {
	Items      []*ProductResponse `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
