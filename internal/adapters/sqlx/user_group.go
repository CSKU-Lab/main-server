package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type userGroupRepository struct {
	db instance
}

type userGroup struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func NewUserGroupRepository(db instance) repositories.UserGroup {
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
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "User not found",
			})
		}
		return nil, err
	}

	return &models.UserGroup{
		ID:   userGroup.ID,
		Name: userGroup.Name,
	}, nil
}

func (r *userGroupRepository) GetByUserID(ctx context.Context, userID string) (string, error) {
	query := `SELECT name
		  FROM user_group_members ugb
		  JOIN user_groups ug
		  ON  ug.id = ugb.group_id
		  WHERE user_id = $1`

	var name string
	err := r.db.GetContext(ctx, &name, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Code:       cserrors.GroupNotFound,
				Message:    "group not found",
			})
		}
		return "", err
	}

	return name, nil

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
