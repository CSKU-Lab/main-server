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
	SetPassword(ctx context.Context, username string, password string) error
	Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error)
	Delete(ctx context.Context, ID string) error
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

func (s *userService) SetPassword(ctx context.Context, username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return fmt.Errorf("Cannot generate password")
	}

	return s.userRepository.SetPassword(ctx, username, string(hashedPassword))
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

	return updatedUser, nil
}

func (s *userService) Delete(ctx context.Context, ID string) error {
	return s.userRepository.Delete(ctx, ID)
}
