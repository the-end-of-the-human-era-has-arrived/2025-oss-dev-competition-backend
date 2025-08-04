package domain

import (
	"github.com/google/uuid"
)

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
