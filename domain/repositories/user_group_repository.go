package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type UserGroup interface {
	Create(ctx context.Context, ID string, name string) error
	GetByID(ctx context.Context, ID string) (*models.UserGroup, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error)
	Count(ctx context.Context, search string) (int, error)
	GetUserAmount(ctx context.Context, ID string) (int, error)
	Update(ctx context.Context, ID string, name string) error
	Delete(ctx context.Context, ID string) error
}
