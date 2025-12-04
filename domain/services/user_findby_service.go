package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type userFindByService struct {
	userRepo          repositories.User
	allowedUserFindBy map[string]bool
}

func NewFindByService(userRepo repositories.User) FindByService[repositories.UserData, *requests.GetInvalidUsers] {
	return &userFindByService{
		userRepo: userRepo,
		allowedUserFindBy: map[string]bool{
			"username": true,
			"email":    true,
		},
	}
}

func (u *userFindByService) Find(ctx context.Context, req *requests.GetInvalidUsers) ([]repositories.UserData, error) {
	err := sanitize.FindBy(u.allowedUserFindBy, req.FindBy)
	if err != nil {
		return nil, err
	}

	users, err := u.userRepo.GetManyByFindBy(
		ctx,
		req.Users,
		req.FindBy,
		req.Role,
	)
	if err != nil {
		return nil, err
	}

	return users, nil
}
