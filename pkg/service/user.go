package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
)

type UserService struct {
	repo *repository.MemoryUserRepo
}

func NewUserService(repo *repository.MemoryUserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	user, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.FindUserByID(ctx, id)
}

func (s *UserService) UpdateNickname(ctx context.Context, id uuid.UUID, nickname string) error {
	s.repo.BeginTransaction(ctx)
	defer s.repo.Commit(ctx)

	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		s.repo.Abort(ctx)
		return err
	}

	user.Nickname = nickname

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		s.repo.Abort(ctx)
		return err
	}
	return nil
}

func (s *UserService) UpdateTokens(
	ctx context.Context,
	id uuid.UUID,
	accessToken, refreshToken string,
) error {
	s.repo.BeginTransaction(ctx)
	defer s.repo.Commit(ctx)

	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		s.repo.Abort(ctx)
		return err
	}

	user.AccessToken = accessToken
	user.RefreshToken = refreshToken

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		s.repo.Abort(ctx)
		return err
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.DeleteUserByID(ctx, id); err != nil {
		return err
	}
	return nil
}
