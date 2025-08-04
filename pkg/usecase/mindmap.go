package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
)

type mindMapUsecase struct {
	repo *repository.MemoryMindMapRepo
}

func NewMindMapUsecase(repo *repository.MemoryMindMapRepo) *mindMapUsecase {
	return &mindMapUsecase{
		repo: repo,
	}
}

func (u *mindMapUsecase) GetMindMapByUser(ctx context.Context, userID uuid.UUID) *domain.MindMapGraph {
	nodes := u.repo.ListKeywordNodeByUser(ctx, userID)

	adjacencyList := make(map[uuid.UUID][]uuid.UUID, len(nodes))
	for _, n := range nodes {
		edges := u.repo.ListKeywordEdgeByNode(ctx, n)
		list := buildNeighborList(n.ID, edges)
		adjacencyList[n.ID] = list
	}

	return &domain.MindMapGraph{
		UserID:        userID,
		AdjacencyList: adjacencyList,
	}
}

func buildNeighborList(id uuid.UUID, edges []*domain.KeywordEdge) []uuid.UUID {
	results := make([]uuid.UUID, 0, len(edges))
	for _, e := range edges {
		neighbor := getNeighborIDFromEdge(id, e)
		results = append(results, neighbor)
	}

	return results
}

func getNeighborIDFromEdge(id uuid.UUID, edge *domain.KeywordEdge) uuid.UUID {
	if edge.Keyword1 == id {
		return edge.Keyword2
	}
	return edge.Keyword1
}
