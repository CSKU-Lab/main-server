package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type userRepository struct {
	db instance
}

func NewUserRepository(db instance) repositories.User {
	return &userRepository{db: db}
}

func buildFilterWhereClause(filters []sanitize.Filter, startingArgIndex int) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []any
	argIndex := startingArgIndex

	for _, filter := range filters {
		switch filter.Operator {
		case "is":
			conditions = append(conditions, fmt.Sprintf("%s = $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " AND " + strings.Join(conditions, " AND "), args
}

type user struct {
	ID           string         `db:"id"`
	Username     string         `db:"username"`
	Type         string         `db:"type"`
	Email        *string        `db:"email"`
	DisplayName  string         `db:"display_name"`
	ProfileImage *string        `db:"profile_image"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	Roles        pq.StringArray `db:"roles"`
	GroupID      *string        `db:"group_id"`
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*repositories.UserData, error) {
	var user user
	query := `SELECT id, email, username,display_name,profile_image,
	roles, group_id, type, created_at, updated_at FROM users WHERE email = $1 AND is_deleted = false`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "User not found"})
		}
		return nil, err
	}

	return &repositories.UserData{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		GroupID:      user.GroupID,
	}, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*repositories.UserData, error) {
	var user user
	query := `SELECT id, email, username,display_name,profile_image,
	roles, group_id, type, created_at, updated_at 
	FROM users 
	WHERE username = $1 AND is_deleted = false`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "User not found"})
		}
		return nil, err
	}

	return &repositories.UserData{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		GroupID:      user.GroupID,
	}, nil
}

func (r *userRepository) GetByID(ctx context.Context, ID string) (*repositories.UserData, error) {
	var user user
	query := `SELECT id, email, username,display_name,profile_image,
	roles, group_id, type, created_at, updated_at 
	FROM users 
	WHERE id = $1 AND is_deleted = false`

	err := r.db.GetContext(ctx, &user, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "User not found"})
		}
		return nil, err
	}

	return &repositories.UserData{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		GroupID:      user.GroupID,
	}, nil
}

func (r *userRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]repositories.UserData, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT id, email, username,display_name,profile_image,
		roles, group_id, type, created_at, updated_at
		FROM users
		WHERE ( username ILIKE $1
		OR display_name ILIKE $1
		OR email ILIKE $1 )
		AND deleted_at IS NULL`

	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+2, len(filterArgs)+3)

	args := []interface{}{"%"+search+"%"}
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	pgUsers := []user{}
	err := r.db.SelectContext(ctx, &pgUsers, query, args...)
	if err != nil {
		return nil, err
	}

	users := make([]repositories.UserData, len(pgUsers))
	for i, pgUser := range pgUsers {
		users[i] = repositories.UserData{
			ID:           pgUser.ID,
			Email:        pgUser.Email,
			Username:     pgUser.Username,
			DisplayName:  pgUser.DisplayName,
			ProfileImage: pgUser.ProfileImage,
			Roles:        pgUser.Roles,
			Type:         pgUser.Type,
			CreatedAt:    pgUser.CreatedAt,
			UpdatedAt:    pgUser.UpdatedAt,
			GroupID:      pgUser.GroupID,
		}
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT COUNT(*) FROM users
		WHERE (username LIKE $1
		OR display_name LIKE $1
		OR email LIKE $1) AND
	        deleted_at IS NULL`

	query := fmt.Sprintf(`%s%s`, baseQuery, filterWhereClause)

	args := []interface{}{"%"+search+"%"}
	args = append(args, filterArgs...)

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *userRepository) Create(ctx context.Context, req repositories.CreateMultiTypeUser) error {
	pgUser := user{
		ID:          req.ID,
		Username:    req.Username,
		Type:        string(req.Type),
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Roles:       req.Roles,
		GroupID:     req.GroupID,
	}

	query, args, err := sqlx.Named(`INSERT INTO users (
			id,
			username,
			display_name,
			email,
			roles,
			type,
			group_id
			) VALUES (:id,:username,:display_name,:email,:roles,:type,:group_id)`, pgUser)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)

	var id string
	err = r.db.GetContext(ctx, &id, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "User already exists"})
			}
		}
	}

	return nil
}

func (r *userRepository) Update(ctx context.Context, ID string, req *requests.UpdateUser) error {
	fields := &user{
		ID:           ID,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Roles:        pq.StringArray(req.Roles),
		Email:        req.Email,
		ProfileImage: req.ProfileImage,
		GroupID:      req.GroupID,
	}

	updateFields := getUpdateFields(fields)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE users
	SET %s, updated_at = NOW()
	WHERE id = :id
	RETURNING id, email, username,display_name,profile_image,
	roles, group_id, type, created_at, updated_at
	`, updateFields)

	query, args, err := sqlx.Named(query, fields)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)

	var updatedUser user
	err = r.db.GetContext(ctx, &updatedUser, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Invalid input for update"})
			}
		}
		return err
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, ID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET is_deleted = true, deleted_at = NOW() WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) DeleteMany(ctx context.Context, IDs []string) error {
	query, args, err := sqlx.In("UPDATE users SET is_deleted = true, deleted_at = NOW() WHERE id IN (?)", IDs)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
