package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
)

type NotionPageService struct {
	repo *repository.MemoryNotionPageRepo
}

func NewNotionPageService(repo *repository.MemoryNotionPageRepo) *NotionPageService {
	return &NotionPageService{
		repo: repo,
	}
}

func (s *NotionPageService) CreateNotionPage(
	ctx context.Context,
	page *domain.NotionPage,
) (*domain.NotionPage, error) {
	createdPage, err := s.repo.CreateNotionPage(ctx, page)
	if err != nil {
		return nil, err
	}

	return createdPage, nil
}

func (s *NotionPageService) GetAllNotionPagesByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.NotionPage, error) {
	return s.repo.FindAllNotionPagesByUser(ctx, userID)
}

func (s *NotionPageService) GetNotionPageByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotionPage, error) {
	return s.repo.FindNotionPageByID(ctx, id)
}

func (s *NotionPageService) UpdateNotionPage(
	ctx context.Context,
	page *domain.NotionPage,
) (*domain.NotionPage, error) {
	updatedPage, err := s.repo.UpdateNotionPage(ctx, page)
	if err != nil {
		return nil, err
	}

	return updatedPage, nil
}

func (s *NotionPageService) DeleteNotionPageByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.NotionPage, error) {
	deletedPage, err := s.repo.DeleteNotionPageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return deletedPage, nil
}
