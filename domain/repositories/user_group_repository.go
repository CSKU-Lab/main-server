package repositories

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
)

type UserGroupRepository interface {
	Create(ctx context.Context, ID string, name string) error
	GetByID(ctx context.Context, ID string) (*models.UserGroup, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error)
	Count(ctx context.Context, search string) (int, error)
	Update(ctx context.Context, ID string, name string) error
	Delete(ctx context.Context, ID string) error
	AddUserToGroup(ctx context.Context, groupID string, userID string) error
}

type UserGroup struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func (u *UserGroup) ToModel() *models.UserGroup {
	return &models.UserGroup{
		ID:   u.ID,
		Name: u.Name,
	}
}
