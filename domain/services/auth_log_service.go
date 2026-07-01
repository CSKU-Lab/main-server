package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/google/uuid"
)

// Auth log action values. Must match the postgres enum "action".
const (
	AuthLogActionSignIn  = "sign-in"
	AuthLogActionRefresh = "refresh"
)

type authLogService struct {
	authLogRepo repositories.AuthLogRepository
}

type AuthLogService interface {
	RecordSignIn(ctx context.Context, userID string) error
	RecordRefresh(ctx context.Context, userID string) error
}

func NewAuthLogService(authLogRepo repositories.AuthLogRepository) AuthLogService {
	return &authLogService{authLogRepo: authLogRepo}
}

func (s *authLogService) RecordSignIn(ctx context.Context, userID string) error {
	return s.record(ctx, userID, AuthLogActionSignIn)
}

func (s *authLogService) RecordRefresh(ctx context.Context, userID string) error {
	return s.record(ctx, userID, AuthLogActionRefresh)
}

func (s *authLogService) record(ctx context.Context, userID string, action string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	return s.authLogRepo.Create(ctx, id.String(), userID, action)
}
