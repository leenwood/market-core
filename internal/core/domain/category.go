package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	ParentID  *uuid.UUID
	SortOrder int
	Children  []*Category
	CreatedAt time.Time
	UpdatedAt time.Time
}
