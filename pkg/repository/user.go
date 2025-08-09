package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
)

type MemoryUserRepo struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

func NewMemoryUserRepo() *MemoryUserRepo {
	return &MemoryUserRepo{
		users: make(map[uuid.UUID]*domain.User, 1024),
	}
}

func (r *MemoryUserRepo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	user.ID = id
	copied := *user

	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[id] = user

	return &copied, nil
}

func (r *MemoryUserRepo) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}

	return user, nil
}

func (r *MemoryUserRepo) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	copied := *user

	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user

	return &copied, nil
}

func (r *MemoryUserRepo) DeleteUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found id: " + id.String())
	}
	delete(r.users, id)

	return deleted, nil
}
