package services

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/google/uuid"
)

type UserGroupService interface {
	Create(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error)
	Update(ctx context.Context, ID string, name string) (*models.UserGroup, error)
	Delete(ctx context.Context, ID string) error
}

type userGroupService struct {
	repo repositories.UserGroupRepository
}

func NewUserGroupService(repo repositories.UserGroupRepository) *userGroupService {
	return &userGroupService{
		repo: repo,
	}
}

func (u *userGroupService) Create(ctx context.Context, name string) error {
	ID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	return u.repo.Create(ctx, ID.String(), name)
}

func (u *userGroupService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &repositories.UserGroup{})
	if err != nil {
		return nil, err
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, err
	}

	return u.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)
}

func (u *userGroupService) Update(ctx context.Context, ID string, name string) (*models.UserGroup, error) {
	err := u.repo.Update(ctx, ID, name)
	if err != nil {
		return nil, err
	}

	return u.repo.GetByID(ctx, ID)
}

func (u *userGroupService) Delete(ctx context.Context, ID string) error {
	return u.repo.Delete(ctx, ID)
}
