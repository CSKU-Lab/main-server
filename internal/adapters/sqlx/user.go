package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxUserRepository struct {
	db instance
}

func NewSqlxUserRepository(db instance) repositories.UserRepository {
	return &sqlxUserRepository{db: db}
}

type postgresUser struct {
	ID           string         `db:"id"`
	Username     string         `db:"username"`
	Type         string         `db:"type"`
	Email        *string        `db:"email"`
	DisplayName  string         `db:"display_name"`
	ProfileImage *string        `db:"profile_image"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	Roles        pq.StringArray `db:"roles"`
}

func (r *sqlxUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user postgresUser
	query := `SELECT  id, email, username,display_name,profile_image,
	roles, type, created_at, updated_at FROM users WHERE email = $1 AND is_deleted = false`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
		}
		return nil, err
	}

	return &models.User{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (r *sqlxUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user postgresUser
	query := `SELECT id, email, username,display_name,profile_image,
	roles, type, created_at, updated_at 
	FROM users 
	WHERE username = $1 AND is_deleted = false`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
		}
		return nil, err
	}

	return &models.User{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (r *sqlxUserRepository) GetByID(ctx context.Context, ID string) (*models.User, error) {
	var user postgresUser
	query := `SELECT id, email, username,display_name,profile_image,
	roles, type, created_at, updated_at 
	FROM users 
	WHERE id = $1 AND is_deleted = false`

	err := r.db.GetContext(ctx, &user, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
		}
		return nil, err
	}

	return &models.User{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Roles:        user.Roles,
		Type:         user.Type,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (r *sqlxUserRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error) {
	query := fmt.Sprintf(`SELECT id, email, username,display_name,profile_image,
		roles, type, created_at, updated_at
		FROM users 
		WHERE username ILIKE $1
		OR display_name ILIKE $1
		OR email ILIKE $1
		AND deleted_at IS NULL
		ORDER BY %s %s
		OFFSET $2
		LIMIT $3
		`, sortBy, sortOrder)

	pgUsers := []postgresUser{}
	err := r.db.SelectContext(ctx, &pgUsers, query, "%"+search+"%", (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, len(pgUsers))
	for i, pgUser := range pgUsers {
		users[i] = models.User{
			ID:           pgUser.ID,
			Email:        pgUser.Email,
			Username:     pgUser.Username,
			DisplayName:  pgUser.DisplayName,
			ProfileImage: pgUser.ProfileImage,
			Roles:        pgUser.Roles,
			Type:         pgUser.Type,
			CreatedAt:    pgUser.CreatedAt,
			UpdatedAt:    pgUser.UpdatedAt,
		}
	}

	return users, nil
}

func (r *sqlxUserRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*) FROM users 
		WHERE (username LIKE $1 
		OR display_name LIKE $1 
		OR email LIKE $1) AND
	        deleted_at IS NULL`

	var count int
	err := r.db.GetContext(ctx, &count, query, "%"+search+"%")
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxUserRepository) Create(ctx context.Context, user repositories.CreateMultiTypeUser) (*models.User, error) {
	pgUser := postgresUser{
		ID:          user.ID,
		Username:    user.Username,
		Type:        user.Type,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       user.Roles,
	}

	query, args, err := sqlx.Named(`INSERT INTO users (
			id,
			username,
			display_name,
			email,
			roles,
			type
			) VALUES (:id,:username,:display_name,:email,:roles,:type)
			RETURNING id, email, username,display_name,profile_image,
		roles, type, created_at, updated_at`, pgUser)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var dbUser postgresUser
	err = r.db.GetContext(ctx, &dbUser, query, args...)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		Username:     dbUser.Username,
		DisplayName:  dbUser.DisplayName,
		ProfileImage: dbUser.ProfileImage,
		Roles:        dbUser.Roles,
		Type:         dbUser.Type,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *sqlxUserRepository) Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error) {
	fields := &postgresUser{
		ID:           ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Roles:        pq.StringArray(user.Roles),
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
	}

	updateFields := getUpdateFields(fields)
	if len(updateFields) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
	UPDATE users
	SET %s, updated_at = NOW()
	WHERE id = :id
	RETURNING id, email, username,display_name,profile_image,
	roles, type, created_at, updated_at
	`, updateFields)

	query, args, err := sqlx.Named(query, fields)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var updatedUser postgresUser
	err = r.db.GetContext(ctx, &updatedUser, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Invalid input for update")
			}
		}
		return nil, err
	}

	return &models.User{
		ID:           updatedUser.ID,
		Email:        updatedUser.Email,
		Username:     updatedUser.Username,
		DisplayName:  updatedUser.DisplayName,
		ProfileImage: updatedUser.ProfileImage,
		Roles:        updatedUser.Roles,
		Type:         updatedUser.Type,
		CreatedAt:    updatedUser.CreatedAt,
		UpdatedAt:    updatedUser.UpdatedAt,
	}, nil
}

func (r *sqlxUserRepository) Delete(ctx context.Context, ID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET is_deleted = true, deleted_at = NOW() WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *sqlxUserRepository) DeleteMany(ctx context.Context, IDs []string) error {
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
