package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/google/uuid"
)

type UserGroupService interface {
	Create(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error)
	Count(ctx context.Context, search string) (int, error)
	Update(ctx context.Context, ID string, name string) (*models.UserGroup, error)
	Delete(ctx context.Context, ID string) error
}

type userGroupService struct {
	repo repositories.UserGroup
}

func NewUserGroupService(repo repositories.UserGroup) *userGroupService {
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
	sanitizedSortBy, err := sanitizeSortBy(sortBy, []string{"name"})
	if err != nil {
		return nil, err
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, err
	}

	return u.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)
}

func (u *userGroupService) Count(ctx context.Context, search string) (int, error) {
	return u.repo.Count(ctx, search)
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
