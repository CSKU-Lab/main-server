package repositories

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, ID string) (*models.User, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error)
	Count(ctx context.Context, search string) (int, error)
	Create(ctx context.Context, user CreateMultiTypeUser) (*models.User, error)
	Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error)
	Delete(ctx context.Context, ID string) error
	DeleteMany(ctx context.Context, IDs []string) error
}

type CreateMultiTypeUser struct {
	requests.CreateMultiTypeUser
	ID string
}
