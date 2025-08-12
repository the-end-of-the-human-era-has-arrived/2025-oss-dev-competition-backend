package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
)

type MemoryNotionPageRepo struct {
	mu     sync.RWMutex
	pages  map[uuid.UUID]*domain.NotionPage
	caches map[uuid.UUID]*notionPageCache
}

type notionPageCache struct {
	mu               sync.RWMutex
	deferedOperation []func()
	pages            map[uuid.UUID]*domain.NotionPage
}

func NewMemoryNotionPageRepo() *MemoryNotionPageRepo {
	return &MemoryNotionPageRepo{
		pages:  make(map[uuid.UUID]*domain.NotionPage, 1024),
		caches: make(map[uuid.UUID]*notionPageCache, 0),
	}
}

func (r *MemoryNotionPageRepo) BeginTransaction(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	pages := make(map[uuid.UUID]*domain.NotionPage)
	for k, v := range r.pages {
		pages[k] = v
	}

	r.caches[requestID] = &notionPageCache{
		deferedOperation: make([]func(), 0),
		pages:            pages,
	}
}

func (r *MemoryNotionPageRepo) Commit(ctx context.Context) {
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

func (r *MemoryNotionPageRepo) Abort(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caches, requestID)
}

func (r *MemoryNotionPageRepo) CreateNotionPage(
	ctx context.Context,
	page *domain.NotionPage,
) (*domain.NotionPage, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	page.ID = id
	copied := *page

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.pages[id] = page
		return &copied, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.pages[id] = page
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.pages[id] = page
	})
	return &copied, nil
}

func (r *MemoryNotionPageRepo) FindAllNotionPagesByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.NotionPage, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		pages := make([]*domain.NotionPage, 0)
		for _, p := range r.pages {
			if p.UserID == userID {
				pages = append(pages, p)
			}
		}
		return pages, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	pages := make([]*domain.NotionPage, 0)
	for _, p := range cache.pages {
		if p.UserID == userID {
			pages = append(pages, p)
		}
	}
	return pages, nil
}

func (r *MemoryNotionPageRepo) FindNotionPageByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotionPage, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		page, ok := r.pages[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		return page, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	page, ok := cache.pages[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	return page, nil
}

func (r *MemoryNotionPageRepo) UpdateNotionPage(
	ctx context.Context,
	page *domain.NotionPage,
) (*domain.NotionPage, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	copied := *page

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.pages[page.ID] = page
		return &copied, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.pages[page.ID] = page
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.pages[page.ID] = page
	})
	return &copied, nil
}

func (r *MemoryNotionPageRepo) DeleteNotionPageByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotionPage, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		deleted, ok := r.pages[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		delete(r.pages, id)
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted, ok := cache.pages[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	delete(cache.pages, id)
	cache.deferedOperation = append(cache.deferedOperation, func() {
		delete(r.pages, id)
	})
	return deleted, nil
}
