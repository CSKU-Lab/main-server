package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, ID string) (*models.User, error)
	GetPasswordByID(ctx context.Context, ID string) (string, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error)
	Count(ctx context.Context, search string) (int, error)
	Create(ctx context.Context, userType models.UserType, user *requests.CreateUser) (*models.User, error)
	CreateMany(ctx context.Context, users *requests.CreateManyUsers) ([]models.User, error)
	SetPassword(ctx context.Context, ID string, password string) error
	Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error)
	Delete(ctx context.Context, ID string) error
	DeleteMany(ctx context.Context, IDs []string) error
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepository.GetByEmail(ctx, email)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepository.GetByUsername(ctx, username)
}

func (s *userService) GetByID(ctx context.Context, ID string) (*models.User, error) {
	return s.userRepository.GetByID(ctx, ID)
}

func (s *userService) GetPasswordByID(ctx context.Context, ID string) (string, error) {
	return s.userRepository.GetPasswordByID(ctx, ID)
}

func (s *userService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &models.User{})
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort by field")
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort order")
	}

	return s.userRepository.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)

}

func (s *userService) Count(ctx context.Context, search string) (int, error) {
	return s.userRepository.Count(ctx, search)
}

func (s *userService) Create(ctx context.Context, userType models.UserType, user *requests.CreateUser) (*models.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate user ID")
	}

	return s.userRepository.Create(ctx, userType, id.String(), user)
}

func (s *userService) CreateMany(ctx context.Context, users *requests.CreateManyUsers) ([]models.User, error) {
	if len(users.Users) == 0 {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "No users to create")
	}

	usersWithIds := make([]requests.CreateMultiTypeUser, 0, 10)
	for _, user := range users.Users {
		if user.ID == "" {
			id, err := uuid.NewV7()
			if err != nil {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate user ID")
			}
			user.ID = id.String()
		}
		usersWithIds = append(usersWithIds, user)
	}

	return s.userRepository.CreateMany(ctx, usersWithIds)
}

func (s *userService) SetPassword(ctx context.Context, ID string, password string) error {
	// TODO: make reuseable function for password hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return fmt.Errorf("Cannot generate password")
	}

	return s.userRepository.SetPassword(ctx, ID, string(hashedPassword))
}

func (s *userService) Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error) {
	dbUser, err := s.userRepository.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	if dbUser.Type == string(models.UserTypeCredential) {
		if user.Email != nil {
			return nil, errors.New("credential user cannot update email")
		}
	}

	updatedUser, err := s.userRepository.Update(ctx, ID, user)
	if err != nil {
		return nil, err
	}

	if dbUser.Type == string(models.UserTypeOauth) {
		if user.Password != nil {
			return nil, errors.New("oauth user cannot update password")
		}
	}

	if user.Password != nil {
		err := s.SetPassword(ctx, ID, *user.Password)
		if err != nil {
			return nil, err
		}
	}

	if updatedUser == nil {
		return dbUser, nil
	}

	return updatedUser, nil
}

func (s *userService) Delete(ctx context.Context, ID string) error {
	return s.userRepository.Delete(ctx, ID)
}

func (s *userService) DeleteMany(ctx context.Context, IDs []string) error {
	if len(IDs) == 0 {
		return nil
	}

	return s.userRepository.DeleteMany(ctx, IDs)
}
