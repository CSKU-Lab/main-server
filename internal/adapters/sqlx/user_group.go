package sqlx

import (
	"context"
	"errors"
	"net/http"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/lib/pq"
)

type userGroupRepository struct {
	db instance
}

type userGroup struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func NewUserGroupRepository(db instance) repositories.UserGroupRepository {
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
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusConflict, Message: "user group already exists"})
			}
		}
		return err
	}
	return nil
}

func (r *userGroupRepository) GetByID(ctx context.Context, ID string) (*models.UserGroup, error) {
	query := `SELECT * FROM user_groups WHERE id = $1`
	var userGroup userGroup
	err := r.db.GetContext(ctx, &userGroup, query, ID)
	if err != nil {
		return nil, err
	}

	return &models.UserGroup{
		ID:   userGroup.ID,
		Name: userGroup.Name,
	}, nil
}

func (r *userGroupRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.UserGroup, error) {
	query := `SELECT * FROM user_groups WHERE name ILIKE $1 ORDER BY ` + sortBy + ` ` + sortOrder + ` LIMIT $2 OFFSET $3`
	searchPattern := "%" + search + "%"
	offset := (page - 1) * limit

	var userGroups []userGroup
	err := r.db.SelectContext(ctx, &userGroups, query, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}

	userGroupModels := make([]models.UserGroup, 0, len(userGroups))
	for _, userGroup := range userGroups {
		userGroupModels = append(userGroupModels, models.UserGroup{
			ID:   userGroup.ID,
			Name: userGroup.Name,
		})
	}

	return userGroupModels, nil
}

func (r *userGroupRepository) Count(ctx context.Context, search string) (int, error) {
	query := `SELECT COUNT(*) FROM user_groups WHERE name ILIKE $1`
	searchPattern := "%" + search + "%"

	var count int
	err := r.db.GetContext(ctx, &count, query, searchPattern)
	if err != nil {
		return 0, err
	}

	return count, nil
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

func (r *userGroupRepository) AddUserToGroup(ctx context.Context, groupID string, userID string) error {
	query := `INSERT INTO user_group_members (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusConflict, Message: "user already in group"})
			}
		}
		return err
	}
	return nil
}
