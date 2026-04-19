package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type TagRepository interface {
	GetPagination(ctx context.Context, page int, limit int, search string) ([]models.Tag, error)
	Count(ctx context.Context, search string) (int, error)
}
