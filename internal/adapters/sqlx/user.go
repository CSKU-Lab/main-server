package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
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

func (r *userRepository) GetManyByFindBy(ctx context.Context, data []string, findBy string, role string) ([]repositories.UserData, error) {
	query := fmt.Sprintf(`
		SELECT id, email, username, display_name, profile_image,
		       roles, group_id, type, created_at, updated_at
		FROM users
		WHERE %s IN (?)
		  AND is_deleted = false
		  AND ? = ANY(roles)
	`, findBy)

	query, args, err := sqlx.In(query, data, role)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var users []user
	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	userDatas := make([]repositories.UserData, len(users))
	for i, u := range users {
		userDatas[i] = repositories.UserData{
			ID:           u.ID,
			Email:        u.Email,
			Username:     u.Username,
			DisplayName:  u.DisplayName,
			ProfileImage: u.ProfileImage,
			Roles:        u.Roles,
			Type:         u.Type,
			CreatedAt:    u.CreatedAt,
			UpdatedAt:    u.UpdatedAt,
			GroupID:      u.GroupID,
		}
	}

	return userDatas, nil
}

func (r *userRepository) GetManyByUsername(ctx context.Context, usernames []string) ([]repositories.UserData, error) {
	var users []user
	query := `SELECT id, email, username,display_name,profile_image,
	roles, group_id, type, created_at, updated_at 
	FROM users 
	WHERE username IN (?) AND is_deleted = false`

	query, args, err := sqlx.In(query, usernames)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "User not found"})
		}
		return nil, err
	}

	var userDatas []repositories.UserData
	for _, user := range users {
		userDatas = append(userDatas, repositories.UserData{
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
		})
	}

	return userDatas, nil
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

	args := []any{"%" + search + "%"}
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

	args := []any{"%" + search + "%"}
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

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusConflict, Message: "User already exists"})
			}
		}
		return err
	}

	return nil
}

func (r *userRepository) Upsert(ctx context.Context, req repositories.CreateMultiTypeUser) error {
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
			) VALUES (:id,:username,:display_name,:email,:roles,:type,:group_id)
			ON CONFLICT (username) WHERE is_deleted = false
			DO UPDATE SET
				display_name = EXCLUDED.display_name,
				email = EXCLUDED.email,
				roles = EXCLUDED.roles,
				group_id = EXCLUDED.group_id,
				updated_at = CURRENT_TIMESTAMP`, pgUser)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
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
