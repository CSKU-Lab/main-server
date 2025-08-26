package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
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
	userRepository         repositories.User
	userPasswordRepository repositories.UserPassword
	userGroupRepository    repositories.UserGroup
	uowRepository          repositories.UserUoWRepository
}

func NewUserService(user repositories.User, userPassword repositories.UserPassword, userGroup repositories.UserGroup, uow repositories.UserUoWRepository) UserService {
	return &userService{
		userRepository:         user,
		userPasswordRepository: userPassword,
		uowRepository:          uow,
		userGroupRepository:    userGroup,
	}
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	dbUser, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	user := dbUser.Model()

	if dbUser.GroupID != nil {
		group, err := s.userGroupRepository.GetByID(ctx, *dbUser.GroupID)
		if err != nil {
			return nil, err
		}
		user.Group = group
	}

	return user, nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	dbUser, err := s.userRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	user := dbUser.Model()

	if dbUser.GroupID != nil {
		group, err := s.userGroupRepository.GetByID(ctx, *dbUser.GroupID)
		if err != nil {
			return nil, err
		}
		user.Group = group
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, ID string) (*models.User, error) {
	dbUser, err := s.userRepository.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	user := dbUser.Model()

	if dbUser.GroupID != nil {
		group, err := s.userGroupRepository.GetByID(ctx, *dbUser.GroupID)
		if err != nil {
			return nil, err
		}
		user.Group = group
	}

	return user, nil
}

func (s *userService) GetPasswordByID(ctx context.Context, ID string) (string, error) {
	return s.userPasswordRepository.GetPasswordByID(ctx, ID)
}

func (s *userService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &models.User{})
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort by field",
		})
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort order",
		})
	}

	dbUsers, err := s.userRepository.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		user := dbUser.Model()

		if dbUser.GroupID != nil {
			group, err := s.userGroupRepository.GetByID(ctx, *dbUser.GroupID)
			if err != nil {
				return nil, err
			}
			user.Group = group
		}
		users[i] = *user
	}

	return users, nil
}

func (s *userService) Count(ctx context.Context, search string) (int, error) {
	return s.userRepository.Count(ctx, search)
}

func (s *userService) Create(ctx context.Context, req *requests.CreateMultiTypeUser) (*models.User, error) {
	if models.UserType(req.Type) == models.UserTypeCredential && req.Email != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Credential user cannot have email",
		})
	}

	if models.UserType(req.Type) == models.UserTypeOauth && req.Password != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Oauth user cannot have password",
		})
	}

	if models.UserType(req.Type) == models.UserTypeCredential && req.Password == nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Credential user must have password",
		})
	}

	if models.UserType(req.Type) == models.UserTypeCredential && req.GroupID == nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Credential user must have group",
		})
	}

	if models.UserType(req.Type) == models.UserTypeCredential {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 10)
		if err != nil {
			return nil, fmt.Errorf("Cannot generate password")
		}
		*req.Password = string(hashedPassword)
	}

	repoUser := repositories.CreateMultiTypeUser{
		CreateMultiTypeUser: *req,
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate user ID",
		})
	}

	repoUser.ID = id.String()

	var user *models.User
	var groupID *string
	err = s.uowRepository.Execute(func(u repositories.UserUoWInstance) error {
		dbUser, err := u.User().Create(ctx, repoUser)
		if err != nil {
			return err
		}

		groupID = dbUser.GroupID
		user = dbUser.Model()

		if req.Password != nil {
			u.UserPassword().SetPassword(ctx, user.ID, *req.Password)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if groupID != nil {
		group, err := s.userGroupRepository.GetByID(ctx, *groupID)
		if err != nil {
			return nil, err
		}

		user.Group = group
	}

	return user, nil
}

func (s *userService) CreateMany(ctx context.Context, req *requests.CreateManyUsers) ([]models.User, error) {
	if len(req.Users) == 0 {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "No users to create",
		})
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

func (s *userService) Update(ctx context.Context, ID string, req *requests.UpdateUser) (*models.User, error) {
	dbUser, err := s.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	if dbUser.Type == string(models.UserTypeCredential) {
		if req.Email != nil {
			return nil, errors.New("credential user cannot update email")
		}
	}

	if dbUser.Type == string(models.UserTypeOauth) {
		if req.Password != nil {
			return nil, errors.New("oauth user cannot update password")
		}
	}

	if req.Password != nil {
		err := s.SetPassword(ctx, ID, *req.Password)
		if err != nil {
			return nil, err
		}
	}

	updatedUser, err := s.userRepository.Update(ctx, ID, req)
	if err != nil {
		return nil, err
	}

	// in case no updated Fields
	if updatedUser == nil {
		return dbUser, nil
	}

	user := updatedUser.Model()

	if updatedUser.GroupID != nil {
		group, err := s.userGroupRepository.GetByID(ctx, *updatedUser.GroupID)
		if err != nil {
			return nil, err
		}

		user.Group = group
	}

	return user, nil
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
