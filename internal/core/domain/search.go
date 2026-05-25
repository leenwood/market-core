package domain

import (
	"time"

	"github.com/google/uuid"
)

type SearchQuery struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	Query        string
	ResultsCount int
	CreatedAt    time.Time
}

type SearchClick struct {
	ID            uuid.UUID
	SearchQueryID uuid.UUID
	ProductID     uuid.UUID
	CreatedAt     time.Time
}

type PopularQuery struct {
	Query string
	Count int64
}

type SearchHistory struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Query     string
	CreatedAt time.Time
}
