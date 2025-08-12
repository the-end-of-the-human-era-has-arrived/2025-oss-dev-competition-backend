package domain

import (
	"github.com/google/uuid"
)

type KeywordNode struct {
	ID           uuid.UUID `json:"id,omitempty"`
	UserID       uuid.UUID `json:"user_id,omitempty"`
	NotionPageID uuid.UUID `json:"notion_page_id,omitempty"`
	Keyword      string    `json:"keyword"`
}

type KeywordEdge struct {
	ID       uuid.UUID `json:"id,omitempty"`
	UserID   uuid.UUID `json:"user_id,omitempty"`
	Keyword1 uuid.UUID `json:"keyword1"`
	Keyword2 uuid.UUID `json:"keyword2"`
}

type EdgeOfIndex struct {
	Idx1 int `json:"idx1"`
	Idx2 int `json:"idx2"`
}

type MindMapGraph struct {
	UserID uuid.UUID      `json:"user_id"`
	Nodes  []*KeywordNode `json:"nodes"`
	Edges  []*KeywordEdge `json:"edges"`
}

func ExtractIDFromBulkNodes(nodes []*KeywordNode) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}

	return ids
}

func ExtractIDFromBulkEdges(edges []*KeywordEdge) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(edges))
	for _, e := range edges {
		ids = append(ids, e.ID)
	}

	return ids
}
