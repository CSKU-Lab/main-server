package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type userActivityService struct {
	userActivityRepo repositories.UserActivityRepository
}

type UserActivityService interface {
	Touch(ctx context.Context, userID string) error
}

func NewUserActivityService(userActivityRepo repositories.UserActivityRepository) UserActivityService {
	return &userActivityService{userActivityRepo: userActivityRepo}
}

func (s *userActivityService) Touch(ctx context.Context, userID string) error {
	return s.userActivityRepo.Touch(ctx, userID)
}
