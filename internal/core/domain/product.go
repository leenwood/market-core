package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	CategoryID  uuid.UUID
	Brand       string
	Price       float64
	Rating      float64
	RatingCount int
	InStock     bool
	Attributes  map[string]any
	ViewCount   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
