package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type userAuthProviderRepository struct {
	db instance
}

func NewUserAuthProviderRepository(db instance) repositories.UserAuthProviderRepository {
	return &userAuthProviderRepository{db: db}
}

func (r *userAuthProviderRepository) GetProviders(ctx context.Context, userID string) ([]models.AuthProvider, error) {
	var providers []string
	err := r.db.SelectContext(ctx, &providers,
		`SELECT provider FROM user_auth_providers WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}

	result := make([]models.AuthProvider, len(providers))
	for i, p := range providers {
		result[i] = models.AuthProvider(p)
	}
	return result, nil
}

func (r *userAuthProviderRepository) HasProvider(ctx context.Context, userID string, provider models.AuthProvider) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM user_auth_providers WHERE user_id = $1 AND provider = $2)`,
		userID, string(provider))
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *userAuthProviderRepository) AddProvider(ctx context.Context, userID string, provider models.AuthProvider, providerID *string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_auth_providers (user_id, provider, provider_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, provider) DO NOTHING`,
		userID, string(provider), providerID)
	return err
}

func (r *userAuthProviderRepository) RemoveProvider(ctx context.Context, userID string, provider models.AuthProvider) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_auth_providers WHERE user_id = $1 AND provider = $2`,
		userID, string(provider))
	return err
}

func (r *userAuthProviderRepository) GetProvidersByUserIDs(ctx context.Context, userIDs []string) (map[string][]models.AuthProvider, error) {
	if len(userIDs) == 0 {
		return map[string][]models.AuthProvider{}, nil
	}

	type row struct {
		UserID   string `db:"user_id"`
		Provider string `db:"provider"`
	}

	query, args, err := sqlx.In(
		`SELECT user_id, provider FROM user_auth_providers WHERE user_id IN (?)`,
		userIDs,
	)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string][]models.AuthProvider)
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], models.AuthProvider(row.Provider))
	}
	return result, nil
}
