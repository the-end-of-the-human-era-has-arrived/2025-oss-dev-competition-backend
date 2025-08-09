package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
)

type MindMapService struct {
	repo *repository.MemoryMindMapRepo
}

func NewMindMapService(repo *repository.MemoryMindMapRepo) *MindMapService {
	return &MindMapService{
		repo: repo,
	}
}

func (s *MindMapService) BuildMindMap(
	ctx context.Context,
	userID uuid.UUID,
	nodes []*domain.KeywordNode,
	edges []*domain.EdgeOfIndex,
) error {
	s.repo.BeginTransaction(ctx)
	defer s.repo.Commit(ctx)

	for _, n := range nodes {
		n.UserID = userID
	}
	newNodes, err := s.repo.CreateBulkKeywordNodes(ctx, nodes...)
	if err != nil {
		s.repo.Abort(ctx)
		return err
	}

	keywordEdges := make([]*domain.KeywordEdge, 0, len(edges))
	for _, e := range edges {
		keywordEdges = append(keywordEdges, &domain.KeywordEdge{
			UserID:   userID,
			Keyword1: newNodes[e.Idx1].ID,
			Keyword2: newNodes[e.Idx2].ID,
		})
	}
	if _, err := s.repo.CreateBulkKeywordEdges(ctx, keywordEdges...); err != nil {
		s.repo.Abort(ctx)
		return err
	}

	return nil
}

func (s *MindMapService) GetMindMapByUser(
	ctx context.Context,
	userID uuid.UUID,
) *domain.MindMapGraph {
	mindmap := &domain.MindMapGraph{
		UserID: userID,
	}

	nodes, err := s.repo.ListKeywordNodeByUser(ctx, userID)
	if err != nil {
		return mindmap
	}
	edges, err := s.repo.ListKeywordEdgeByUser(ctx, userID)
	if err != nil {
		return mindmap
	}

	mindmap.Nodes = nodes
	mindmap.Edges = edges
	return mindmap
}

func (s *MindMapService) DeleteMindMapByUser(ctx context.Context, userID uuid.UUID) error {
	s.repo.BeginTransaction(ctx)
	defer s.repo.Commit(ctx)

	nodes, err := s.repo.ListKeywordNodeByUser(ctx, userID)
	if err != nil {
		s.repo.Abort(ctx)
		return err
	}

	ids := domain.ExtractIDFromBulkNodes(nodes)
	if _, err := s.repo.DeleteBulkKeywordNodes(ctx, ids); err != nil {
		s.repo.Abort(ctx)
		return err
	}

	edges, err := s.repo.ListKeywordEdgeByUser(ctx, userID)
	if err != nil {
		s.repo.Abort(ctx)
		return err
	}

	ids = domain.ExtractIDFromBulkEdges(edges)
	if _, err := s.repo.DeleteBulkKeywordEdges(ctx, ids); err != nil {
		s.repo.Abort(ctx)
		return err
	}

	return nil
}
