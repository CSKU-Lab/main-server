package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/converter"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type User interface {
	GetByEmail(ctx context.Context, email string) (*UserData, error)
	GetByUsername(ctx context.Context, username string) (*UserData, error)
	GetByID(ctx context.Context, ID string) (*UserData, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]UserData, error)
	Count(ctx context.Context, search string) (int, error)
	Create(ctx context.Context, user CreateMultiTypeUser) (*UserData, error)
	Update(ctx context.Context, ID string, user *requests.UpdateUser) (*UserData, error)
	Delete(ctx context.Context, ID string) error
	DeleteMany(ctx context.Context, IDs []string) error
}

type CreateMultiTypeUser struct {
	requests.CreateMultiTypeUser
	ID string
}

type UserData struct {
	ID           string
	Username     string
	Type         string
	Email        *string
	DisplayName  string
	ProfileImage *string
	Roles        []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	GroupID      *string
}

func (u *UserData) Model() (*models.User, error) {
	userRoles, err := converter.ToRoleSlice(u.Roles)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           u.ID,
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		Type:         u.Type,
		Email:        u.Email,
		ProfileImage: u.ProfileImage,
		Roles:        userRoles,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.CreatedAt,
	}, nil
}
