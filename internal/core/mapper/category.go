package mapper

import (
	"time"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"

	"github.com/google/uuid"
)

func CategoryToResponse(c *domain.Category) *dto.CategoryResponse {
	resp := &dto.CategoryResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		Slug:      c.Slug,
		SortOrder: c.SortOrder,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		resp.ParentID = &s
	}
	if len(c.Children) > 0 {
		resp.Children = make([]*dto.CategoryResponse, len(c.Children))
		for i, ch := range c.Children {
			resp.Children[i] = CategoryToResponse(ch)
		}
	}
	return resp
}

func CategoriesToResponse(categories []*domain.Category) []*dto.CategoryResponse {
	out := make([]*dto.CategoryResponse, len(categories))
	for i, c := range categories {
		out[i] = CategoryToResponse(c)
	}
	return out
}

func CreateRequestToCategory(req dto.CreateCategoryRequest) *domain.Category {
	return &domain.Category{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
