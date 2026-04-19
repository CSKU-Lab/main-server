package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type TagService interface {
	GetPagination(ctx context.Context, page int, limit int, search string) ([]models.Tag, error)
	Count(ctx context.Context, search string) (int, error)
}

type tagService struct {
	repo repositories.TagRepository
}

func NewTagService(repo repositories.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) GetPagination(ctx context.Context, page int, limit int, search string) ([]models.Tag, error) {
	return s.repo.GetPagination(ctx, page, limit, search)
}

func (s *tagService) Count(ctx context.Context, search string) (int, error) {
	return s.repo.Count(ctx, search)
}
