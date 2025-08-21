package sqlx

import (
	"context"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxRefreshTokenRepository struct {
	db *sqlx.DB
}

func NewSQLxRefreshTokenRepository(db *sqlx.DB) repositories.RefreshTokenRepository {
	return &sqlxRefreshTokenRepository{db: db}
}

func (r *sqlxRefreshTokenRepository) Get(ctx context.Context, userID string) (string, error) {
	var token string
	err := r.db.GetContext(ctx, &token, "SELECT token FROM user_refresh_tokens WHERE user_id = $1", userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return "", cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Refresh token not found"})
			}
		}
	}

	return token, nil
}

func (r *sqlxRefreshTokenRepository) Set(ctx context.Context, userID string, token string) error {
	query := `
		INSERT INTO user_refresh_tokens (user_id, token)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE
		SET token = $2
	`
	_, err := r.db.ExecContext(ctx, query, userID, token)
	if err != nil {
		return err
	}
	return nil
}
