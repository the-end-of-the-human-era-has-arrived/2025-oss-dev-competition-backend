package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
)

type MemoryMindMapRepo struct {
	mu     sync.RWMutex
	nodes  map[uuid.UUID]*domain.KeywordNode
	edges  map[uuid.UUID]*domain.KeywordEdge
	caches map[uuid.UUID]*mindmapCache
}

type mindmapCache struct {
	mu               sync.RWMutex
	deferedOperation []func()
	nodes            map[uuid.UUID]*domain.KeywordNode
	edges            map[uuid.UUID]*domain.KeywordEdge
}

func NewMemoryMindMapRepo() *MemoryMindMapRepo {
	return &MemoryMindMapRepo{
		nodes:  make(map[uuid.UUID]*domain.KeywordNode, 1024),
		edges:  make(map[uuid.UUID]*domain.KeywordEdge, 1024),
		caches: make(map[uuid.UUID]*mindmapCache, 0),
	}
}

func (r *MemoryMindMapRepo) BeginTransaction(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := make(map[uuid.UUID]*domain.KeywordNode)
	for k, v := range r.nodes {
		nodes[k] = v
	}

	edges := make(map[uuid.UUID]*domain.KeywordEdge)
	for k, v := range r.edges {
		edges[k] = v
	}

	r.caches[requestID] = &mindmapCache{
		deferedOperation: make([]func(), 0),
		nodes:            nodes,
		edges:            edges,
	}
}

func (r *MemoryMindMapRepo) Commit(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		return
	}
	for _, operation := range cache.deferedOperation {
		operation()
	}
}

func (r *MemoryMindMapRepo) Abort(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caches, requestID)
}

func (r *MemoryMindMapRepo) CreateKeywordNode(
	ctx context.Context,
	node *domain.KeywordNode,
) (*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	node.ID = id
	copied := *node

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.nodes[id] = node
		return &copied, nil
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.nodes[id] = node
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.nodes[id] = node
	})

	return &copied, nil
}

func (r *MemoryMindMapRepo) CreateBulkKeywordNodes(
	ctx context.Context,
	bulks ...*domain.KeywordNode,
) ([]*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	nodes := make([]*domain.KeywordNode, 0, len(bulks))
	for _, node := range bulks {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		node.ID = id
		nodes = append(nodes, node)
	}

	nodeCopies := make([]*domain.KeywordNode, 0, len(nodes))
	for _, node := range nodes {
		copied := *node
		nodeCopies = append(nodeCopies, &copied)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		for _, node := range nodes {
			r.nodes[node.ID] = node
		}
		return nodeCopies, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	for _, node := range nodes {
		cache.nodes[node.ID] = node
	}
	cache.deferedOperation = append(cache.deferedOperation, func() {
		for _, node := range nodes {
			r.nodes[node.ID] = node
		}
	})

	return nodeCopies, nil
}

func (r *MemoryMindMapRepo) FindKeywordNodeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		node, ok := r.nodes[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}

		return node, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	node, ok := cache.nodes[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}

	return node, nil
}

func (r *MemoryMindMapRepo) ListKeywordNodeByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		nodes := make([]*domain.KeywordNode, 0)
		for _, n := range r.nodes {
			if n.UserID == userID {
				nodes = append(nodes, n)
			}
		}
		return nodes, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	nodes := make([]*domain.KeywordNode, 0)
	for _, n := range cache.nodes {
		if n.UserID == userID {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (r *MemoryMindMapRepo) ListKeywordNodeByNotionPage(
	ctx context.Context,
	notionID uuid.UUID,
) ([]*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		nodes := make([]*domain.KeywordNode, 0)
		for _, n := range r.nodes {
			if n.NotionPageID == notionID {
				nodes = append(nodes, n)
			}
		}
		return nodes, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	nodes := make([]*domain.KeywordNode, 0)
	for _, n := range cache.nodes {
		if n.NotionPageID == notionID {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (r *MemoryMindMapRepo) DeleteKeywordNodeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		deleted, ok := r.nodes[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		delete(r.nodes, id)
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted, ok := cache.nodes[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	delete(cache.nodes, id)
	cache.deferedOperation = append(cache.deferedOperation, func() {
		delete(r.nodes, id)
	})

	return deleted, nil
}

func (r *MemoryMindMapRepo) DeleteBulkKeywordNodes(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domain.KeywordNode, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cache, ok := r.caches[requestID]
	if !ok {
		deleted := make([]*domain.KeywordNode, 0, len(ids))
		for _, id := range ids {
			node, ok := r.nodes[id]
			if !ok {
				return nil, errors.New("not found id: " + id.String())
			}
			deleted = append(deleted, node)
		}
		for _, id := range ids {
			delete(r.nodes, id)
		}
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted := make([]*domain.KeywordNode, 0, len(ids))
	for _, id := range ids {
		node, ok := cache.nodes[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		deleted = append(deleted, node)
	}
	for _, id := range ids {
		delete(cache.nodes, id)
	}
	cache.deferedOperation = append(cache.deferedOperation, func() {
		for _, id := range ids {
			delete(r.nodes, id)
		}
	})
	return deleted, nil
}

func (r *MemoryMindMapRepo) CreateKeywordEdge(
	ctx context.Context,
	edge *domain.KeywordEdge,
) (*domain.KeywordEdge, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	edge.ID = id
	copied := *edge

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.edges[id] = edge
		return &copied, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.edges[id] = edge
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.edges[id] = edge
	})

	return &copied, nil
}

func (r *MemoryMindMapRepo) CreateBulkKeywordEdges(
	ctx context.Context,
	bulks ...*domain.KeywordEdge,
) ([]*domain.KeywordEdge, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	edges := make([]*domain.KeywordEdge, 0, len(bulks))
	for _, edge := range bulks {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		edge.ID = id
		edges = append(edges, edge)
	}

	edgeCopies := make([]*domain.KeywordEdge, 0, len(edges))
	for _, edge := range edges {
		copied := *edge
		edgeCopies = append(edgeCopies, &copied)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		for _, edge := range edges {
			r.edges[edge.ID] = edge
		}
		return edgeCopies, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	for _, edge := range edges {
		cache.edges[edge.ID] = edge
	}
	cache.deferedOperation = append(cache.deferedOperation, func() {
		for _, edge := range edges {
			r.edges[edge.ID] = edge
		}
	})

	return edgeCopies, nil
}

func (r *MemoryMindMapRepo) ListKeywordEdgeByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.KeywordEdge, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	cache, ok := r.caches[requestID]
	if !ok {
		edges := make([]*domain.KeywordEdge, 0)
		for _, e := range r.edges {
			if e.UserID == userID {
				edges = append(edges, e)
			}
		}
		return edges, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	edges := make([]*domain.KeywordEdge, 0)
	for _, e := range cache.edges {
		if e.UserID == userID {
			edges = append(edges, e)
		}
	}

	return edges, nil
}

func (r *MemoryMindMapRepo) DeleteKeywordEdgeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.KeywordEdge, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cache, ok := r.caches[requestID]
	if !ok {
		deleted, ok := r.edges[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		delete(r.edges, id)
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted, ok := cache.edges[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	delete(cache.edges, id)
	cache.deferedOperation = append(cache.deferedOperation, func() {
		delete(r.edges, id)
	})

	return deleted, nil
}

func (r *MemoryMindMapRepo) DeleteBulkKeywordEdges(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domain.KeywordEdge, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cache, ok := r.caches[requestID]
	if !ok {
		deleted := make([]*domain.KeywordEdge, 0, len(ids))
		for _, id := range ids {
			node, ok := r.edges[id]
			if !ok {
				return nil, errors.New("not found id: " + id.String())
			}
			deleted = append(deleted, node)
		}
		for _, id := range ids {
			delete(r.edges, id)
		}
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted := make([]*domain.KeywordEdge, 0, len(ids))
	for _, id := range ids {
		node, ok := cache.edges[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		deleted = append(deleted, node)
	}
	for _, id := range ids {
		delete(cache.edges, id)
	}
	cache.deferedOperation = append(cache.deferedOperation, func() {
		for _, id := range ids {
			delete(r.edges, id)
		}
	})

	return deleted, nil
}
