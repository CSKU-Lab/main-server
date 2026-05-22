package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type SearchService interface {
	Search(ctx context.Context, q string, limit int) (*models.SearchResult, error)
}

type searchService struct {
	searchRepo repositories.SearchRepository
}

func NewSearchService(searchRepo repositories.SearchRepository) SearchService {
	return &searchService{searchRepo: searchRepo}
}

func (s *searchService) Search(ctx context.Context, q string, limit int) (*models.SearchResult, error) {
	return s.searchRepo.Search(ctx, q, limit)
}
