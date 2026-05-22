package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SearchRepository interface {
	Search(ctx context.Context, q string, limit int) (*models.SearchResult, error)
}
