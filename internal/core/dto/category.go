package dto

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	Name      string     `json:"name" validate:"required,min=1,max=255"`
	Slug      string     `json:"slug" validate:"required,min=1,max=255"`
	ParentID  *uuid.UUID `json:"parent_id"`
	SortOrder int        `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=255"`
	Slug      *string `json:"slug" validate:"omitempty,min=1,max=255"`
	SortOrder *int    `json:"sort_order"`
}

type CategoryResponse struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Slug      string              `json:"slug"`
	ParentID  *string             `json:"parent_id,omitempty"`
	SortOrder int                 `json:"sort_order"`
	Children  []*CategoryResponse `json:"children,omitempty"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
}
