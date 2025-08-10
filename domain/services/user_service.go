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
	Create(ctx context.Context, user *requests.CreateMultiTypeUser) (*models.User, error)
	CreateMany(ctx context.Context, users *requests.CreateManyUsers) ([]models.User, error)
	SetPassword(ctx context.Context, ID string, password string) error
	Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error)
	Delete(ctx context.Context, ID string) error
	DeleteMany(ctx context.Context, IDs []string) error
}

type userService struct {
	userRepository         repositories.UserRepository
	userPasswordRepository repositories.UserPasswordRepository
	uowRepository          repositories.UserUoWRepository
}

func NewUserService(userRepository repositories.UserRepository, userPasswordRepository repositories.UserPasswordRepository, uowRepository repositories.UserUoWRepository) UserService {
	return &userService{
		userRepository:         userRepository,
		userPasswordRepository: userPasswordRepository,
		uowRepository:          uowRepository,
	}
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
	return s.userPasswordRepository.GetPasswordByID(ctx, ID)
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

func (s *userService) Create(ctx context.Context, req *requests.CreateMultiTypeUser) (*models.User, error) {
	if req.Type == models.UserTypeCredential.String() && req.Email != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Credential user cannot have email")
	}

	if req.Type == models.UserTypeOauth.String() && req.Password != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Oauth user cannot have password")
	}

	if req.Type == models.UserTypeCredential.String() && req.Password == nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Credential user must have password")
	}

	if req.Type == models.UserTypeCredential.String() && req.GroupID == nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Credential user must have group")
	}

	repoUser := repositories.CreateMultiTypeUser{
		CreateMultiTypeUser: *req,
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate user ID")
	}

	repoUser.ID = id.String()

	var user *models.User
	err = s.uowRepository.Execute(func(u repositories.UserUoWInstance) {
		user, _ = u.User().Create(ctx, repoUser)

		if req.GroupID != nil {
			u.UserGroup().AddUserToGroup(ctx, *req.GroupID, user.ID)

			group, _ := u.UserGroup().GetByID(ctx, *req.GroupID)
			user.Group = &group.Name
		}

		if req.Password != nil {
			u.UserPassword().SetPassword(ctx, user.ID, *req.Password)
		}
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) CreateMany(ctx context.Context, req *requests.CreateManyUsers) ([]models.User, error) {
	if len(req.Users) == 0 {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "No users to create")
	}

	users := make([]models.User, len(req.Users))
	for i, user := range req.Users {
		user, err := s.Create(ctx, &user)
		if err != nil {
			return nil, err
		}

		users[i] = *user
	}

	return users, nil
}

func (s *userService) SetPassword(ctx context.Context, ID string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return fmt.Errorf("Cannot generate password")
	}

	return s.userPasswordRepository.SetPassword(ctx, ID, string(hashedPassword))
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
