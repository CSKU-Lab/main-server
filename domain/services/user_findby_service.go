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

type (
	PreferResponse map[string]bool
	PreferRequest  *requests.GetInvalidUsers
)

func NewFindByService(userRepo repositories.User) FindByService[PreferResponse, PreferRequest] {
	return &userFindByService{
		userRepo: userRepo,
		allowedUserFindBy: map[string]bool{
			"username": true,
			"email":    true,
		},
	}
}

func (u *userFindByService) Find(ctx context.Context, req PreferRequest) (PreferResponse, error) {
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

	flattenUsers := map[string]bool{}
	for _, u := range users {
		switch req.FindBy {
		case "username":
			flattenUsers[u.Username] = true
		case "email":
			if u.Email != nil {
				flattenUsers[*u.Email] = true
			}
		}
	}
	return flattenUsers, nil
}
