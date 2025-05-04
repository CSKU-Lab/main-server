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
	GetPasswordByID(ctx context.Context, ID string) (string, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error)
	Count(ctx context.Context, search string) (int, error)
	Create(ctx context.Context, ID string, user *requests.User) (*models.User, error)
	SetPassword(ctx context.Context, username string, password string) error
	Update(ctx context.Context, ID string, user *requests.User) (*models.User, error)
	Delete(ctx context.Context, ID string) error
}
