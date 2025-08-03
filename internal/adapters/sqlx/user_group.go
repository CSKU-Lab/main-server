package sqlx

import (
	"context"
	"errors"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type userGroupRepository struct {
	db *sqlx.DB
}

func NewUserGroupRepository(db *sqlx.DB) repositories.UserGroupRepository {
	return &userGroupRepository{
		db: db,
	}
}

func (r *userGroupRepository) Create(ctx context.Context, ID string, name string) error {
	query := `INSERT INTO user_groups (id, name) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, ID, name)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(cserrors.ALREADY_EXISTS, "user group already exists")
			}
		}
		return err
	}
	return nil
}

func (r *userGroupRepository) GetByID(ctx context.Context, ID string) (*models.UserGroup, error) {
	query := `SELECT * FROM user_groups WHERE id = $1`
	var userGroup repositories.UserGroup
	err := r.db.GetContext(ctx, &userGroup, query, ID)
	if err != nil {
		return nil, err
	}

	return userGroup.ToModel(), nil
}

func (r *userGroupRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error) {
	query := `SELECT * FROM user_groups WHERE name ILIKE $1 ORDER BY ` + sortBy + ` ` + sortOrder + ` LIMIT $2 OFFSET $3`
	searchPattern := "%" + search + "%"
	offset := (page - 1) * limit

	var userGroups []repositories.UserGroup
	err := r.db.SelectContext(ctx, &userGroups, query, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}

	userGroupModels := make([]models.UserGroup, 0, len(userGroups))
	for _, userGroup := range userGroups {
		userGroupModels = append(userGroupModels, *userGroup.ToModel())
	}

	return userGroupModels, nil
}

func (r *userGroupRepository) Update(ctx context.Context, ID string, name string) error {
	query := `UPDATE user_groups SET name = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, name, ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *userGroupRepository) Delete(ctx context.Context, ID string) error {
	query := `DELETE FROM user_groups WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, ID)
	if err != nil {
		return err
	}
	return nil
}
