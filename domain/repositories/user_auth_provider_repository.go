package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type UserAuthProviderRepository interface {
	GetProviders(ctx context.Context, userID string) ([]models.AuthProvider, error)
	GetProvidersByUserIDs(ctx context.Context, userIDs []string) (map[string][]models.AuthProvider, error)
	HasProvider(ctx context.Context, userID string, provider models.AuthProvider) (bool, error)
	AddProvider(ctx context.Context, userID string, provider models.AuthProvider, providerID *string) error
	RemoveProvider(ctx context.Context, userID string, provider models.AuthProvider) error
}
