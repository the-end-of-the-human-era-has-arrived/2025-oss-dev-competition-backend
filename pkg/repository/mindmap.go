package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
)

type MemoryMindMapRepo struct {
	mu    sync.RWMutex
	nodes map[uuid.UUID]*domain.KeywordNode
	edges map[uuid.UUID]*domain.KeywordEdge
}

func NewMemoryMindMapRepo() *MemoryMindMapRepo {
	return &MemoryMindMapRepo{
		nodes: make(map[uuid.UUID]*domain.KeywordNode, 1024),
		edges: make(map[uuid.UUID]*domain.KeywordEdge, 1024),
	}
}

func (r *MemoryMindMapRepo) CreateKeywordNode(
	ctx context.Context,
	userID, notionID uuid.UUID,
	keyword string,
) (*domain.KeywordNode, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	newNode := domain.KeywordNode{
		ID:           id,
		UserID:       userID,
		NotionPageID: notionID,
		Keyword:      keyword,
	}
	copied := newNode

	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[id] = &newNode

	return &copied, nil
}

func (r *MemoryMindMapRepo) FindKeywordNodeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}

	return node, nil
}

func (r *MemoryMindMapRepo) ListKeywordNodeByUser(
	ctx context.Context,
	userID uuid.UUID,
) []*domain.KeywordNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*domain.KeywordNode, 0)
	for _, n := range r.nodes {
		if n.UserID == userID {
			nodes = append(nodes, n)
		}
	}

	return nodes
}

func (r *MemoryMindMapRepo) ListKeywordNodeByNotionPage(
	ctx context.Context,
	notionID uuid.UUID,
) []*domain.KeywordNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*domain.KeywordNode, 0)
	for _, n := range r.nodes {
		if n.NotionPageID == notionID {
			nodes = append(nodes, n)
		}
	}

	return nodes
}

func (r *MemoryMindMapRepo) DeleteKeywordNode(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordNode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted, ok := r.nodes[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}

	delete(r.nodes, id)

	return deleted, nil
}

func (r *MemoryMindMapRepo) CreateKeywordEdge(
	ctx context.Context,
	node1, node2 *domain.KeywordNode,
) (*domain.KeywordEdge, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	newEdge := &domain.KeywordEdge{
		ID:       id,
		Keyword1: node1.ID,
		Keyword2: node2.ID,
	}
	copied := *newEdge

	r.mu.Lock()
	defer r.mu.Unlock()
	r.edges[id] = newEdge

	return &copied, nil
}

func (r *MemoryMindMapRepo) ListKeywordEdgeByNode(
	ctx context.Context,
	node *domain.KeywordNode,
) []*domain.KeywordEdge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	edges := make([]*domain.KeywordEdge, 0)
	for _, e := range r.edges {
		if e.Keyword1 == node.ID || e.Keyword2 == node.ID {
			edges = append(edges, e)
		}
	}

	return edges
}

func (r *MemoryMindMapRepo) DeleteKeywordEdgeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordEdge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted, ok := r.edges[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}

	delete(r.edges, id)

	return deleted, nil
}
