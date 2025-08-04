package domain

import (
	"time"

	"github.com/google/uuid"
)

// Legacy models - keeping as is
type KeywordNode struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	NotionPageID uuid.UUID
	Keyword      string
}

type KeywordEdge struct {
	ID       uuid.UUID
	Keyword1 uuid.UUID
	Keyword2 uuid.UUID
}

type MindMapGraph struct {
	UserID        uuid.UUID                 `json:"userID"`
	AdjacencyList map[uuid.UUID][]uuid.UUID `json:"adjacencyList"`
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
}
