package mapper

import (
	"time"

	"market-core/internal/core/domain"
	"market-core/internal/core/dto"

	"github.com/google/uuid"
)

func ProductToResponse(p *domain.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		CategoryID:  p.CategoryID.String(),
		Brand:       p.Brand,
		Price:       p.Price,
		Rating:      p.Rating,
		RatingCount: p.RatingCount,
		InStock:     p.InStock,
		Attributes:  p.Attributes,
		ViewCount:   p.ViewCount,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func ProductsToResponse(products []*domain.Product) []*dto.ProductResponse {
	out := make([]*dto.ProductResponse, len(products))
	for i, p := range products {
		out[i] = ProductToResponse(p)
	}
	return out
}

func CreateRequestToProduct(req dto.CreateProductRequest) *domain.Product {
	attrs := req.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}
	return &domain.Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		Price:       req.Price,
		InStock:     req.InStock,
		Attributes:  attrs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func ApplyUpdateRequest(p *domain.Product, req dto.UpdateProductRequest) {
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.CategoryID != nil {
		p.CategoryID = *req.CategoryID
	}
	if req.Brand != nil {
		p.Brand = *req.Brand
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.InStock != nil {
		p.InStock = *req.InStock
	}
	if req.Attributes != nil {
		p.Attributes = req.Attributes
	}
	p.UpdatedAt = time.Now()
}
