package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
)

type MemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[uuid.UUID]*domain.User
	caches map[uuid.UUID]*userCache
}

type userCache struct {
	mu               sync.RWMutex
	deferedOperation []func()
	users            map[uuid.UUID]*domain.User
}

func NewMemoryUserRepo() *MemoryUserRepo {
	return &MemoryUserRepo{
		users:  make(map[uuid.UUID]*domain.User, 1024),
		caches: make(map[uuid.UUID]*userCache, 0),
	}
}

func (r *MemoryUserRepo) BeginTransaction(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	users := make(map[uuid.UUID]*domain.User)
	for k, v := range r.users {
		users[k] = v
	}

	r.caches[requestID] = &userCache{
		deferedOperation: make([]func(), 0),
		users:            users,
	}
}

func (r *MemoryUserRepo) Commit(ctx context.Context) {
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

func (r *MemoryUserRepo) Abort(ctx context.Context) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caches, requestID)
}

func (r *MemoryUserRepo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user.ID = id
	copied := *user

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.users[id] = user
		return &copied, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.users[id] = user
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.users[id] = user
	})
	return &copied, nil
}

func (r *MemoryUserRepo) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	cache, ok := r.caches[requestID]
	if !ok {
		user, ok := r.users[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		return user, nil
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()
	user, ok := cache.users[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	return user, nil
}

func (r *MemoryUserRepo) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	copied := *user

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		r.users[user.ID] = user
		return &copied, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.users[user.ID] = user
	cache.deferedOperation = append(cache.deferedOperation, func() {
		r.users[user.ID] = user
	})
	return &copied, nil
}

func (r *MemoryUserRepo) DeleteUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	requestID, ok := ctx.Value(api.RequestIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("not found request id")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	cache, ok := r.caches[requestID]
	if !ok {
		deleted, ok := r.users[id]
		if !ok {
			return nil, errors.New("not found id: " + id.String())
		}
		delete(r.users, id)
		return deleted, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	deleted, ok := cache.users[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	delete(cache.users, id)
	cache.deferedOperation = append(cache.deferedOperation, func() {
		delete(r.users, id)
	})
	return deleted, nil
}
