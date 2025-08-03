package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type sqlxUserRepository struct {
	db *sqlx.DB
}

func NewSqlxUserRepository(db *sqlx.DB) repositories.UserRepository {
	return &sqlxUserRepository{db: db}
}

type PostgresUser struct {
	models.User
	Roles pq.StringArray `db:"roles"`
}

func (r *sqlxUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user PostgresUser
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1 AND is_deleted = false", email)
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
		RecordStatus: models.RecordStatus{
			IsDeleted: user.IsDeleted,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		},
	}, nil
}

func (r *sqlxUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user PostgresUser
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE username = $1 AND is_deleted = false", username)
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
		RecordStatus: models.RecordStatus{
			IsDeleted: user.IsDeleted,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		},
	}, nil
}

func (r *sqlxUserRepository) GetByID(ctx context.Context, ID string) (*models.User, error) {
	var user PostgresUser
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1 AND is_deleted = false", ID)
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
		RecordStatus: models.RecordStatus{
			IsDeleted: user.IsDeleted,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		},
	}, nil
}

func (r *sqlxUserRepository) GetPasswordByID(ctx context.Context, ID string) (string, error) {
	row := r.db.QueryRowContext(ctx, "SELECT password FROM user_passwords WHERE user_id = $1", ID)
	var password string

	err := row.Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
		}
		return "", err
	}

	return password, nil
}

func (r *sqlxUserRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.User, error) {
	query := fmt.Sprintf(`SELECT * FROM users 
		WHERE (LOWER(username) LIKE LOWER($1)
		OR LOWER(display_name) LIKE LOWER($1)
		OR LOWER(email) LIKE LOWER($1))
		AND deleted_at IS NULL
		ORDER BY %s %s
		OFFSET $2
		LIMIT $3
		`, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}

	users := []models.User{}

	for rows.Next() {
		var user PostgresUser
		err = rows.StructScan(&user)
		if err != nil {
			return nil, err
		}

		users = append(users, models.User{
			ID:           user.ID,
			Email:        user.Email,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			ProfileImage: user.ProfileImage,
			Roles:        user.Roles,
			Type:         user.Type,
			RecordStatus: models.RecordStatus{
				IsDeleted: user.IsDeleted,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				DeletedAt: user.DeletedAt,
			},
		})
	}

	return users, nil
}

func (r *sqlxUserRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*) FROM users 
		WHERE (username LIKE $1 
		OR display_name LIKE $1 
		OR email LIKE $1) AND
	        deleted_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, "%"+search+"%")

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxUserRepository) Create(ctx context.Context, userType models.UserType, ID string, user *requests.CreateUser) (*models.User, error) {
	createString := `
		INSERT INTO users (
			id,
			username,
			display_name,
			email,
			roles,
			type
		) VALUES ($1,$2,$3,$4,string_to_array($5,',')::role[],$6)
		RETURNING *
	`

	User := r.db.QueryRowxContext(ctx, createString, ID, user.Username, user.DisplayName, user.Email, strings.Join(user.Roles, ","), userType)

	var createdUser PostgresUser

	err := User.StructScan(&createdUser)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User already exists")
			}
		}
		return nil, err
	}

	return &models.User{
		ID:           createdUser.ID,
		Email:        createdUser.Email,
		Username:     createdUser.Username,
		DisplayName:  createdUser.DisplayName,
		ProfileImage: createdUser.ProfileImage,
		Roles:        createdUser.Roles,
		Type:         createdUser.Type,
		RecordStatus: models.RecordStatus{
			IsDeleted: createdUser.IsDeleted,
			CreatedAt: createdUser.CreatedAt,
			UpdatedAt: createdUser.UpdatedAt,
			DeletedAt: createdUser.DeletedAt,
		},
	}, nil
}

func (r *sqlxUserRepository) CreateMany(ctx context.Context, users []requests.CreateMultiTypeUser) ([]models.User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot begin transaction")
	}

	var insertedUsers []models.User
	for _, user := range users {
		pgUser := map[string]any{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"email":        user.Email,
			"roles":        pq.StringArray(user.Roles),
			"type":         user.Type,
		}

		query, args, err := sqlx.Named(`INSERT INTO users (
			id,
			username,
			display_name,
			email,
			roles,
			type
			) VALUES (:id,:username,:display_name,:email,:roles,:type)
			RETURNING *`, pgUser)
		if err != nil {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "sqlx.Named")
		}

		query = tx.Rebind(query)
		insertedUser := tx.QueryRowx(query, args...)

		var dbUser PostgresUser
		err = insertedUser.StructScan(&dbUser)
		if err != nil {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, err.Error())
		}

		if user.Type == string(models.UserTypeCredential) && user.Password != nil {
			// TODO: make reuseable function for password hashing
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), 10)
			if err != nil {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate password")
			}

			_, err = tx.ExecContext(ctx, "INSERT INTO user_passwords (user_id, password) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET password = $2", dbUser.ID, string(hashedPassword))
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction: %v", rollbackErr)
					return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot rollback transaction")
				}
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot set password")
			}
		}

		insertedUsers = append(insertedUsers, models.User{
			ID:           dbUser.ID,
			Email:        dbUser.Email,
			Username:     dbUser.Username,
			DisplayName:  dbUser.DisplayName,
			ProfileImage: dbUser.ProfileImage,
			Roles:        user.Roles,
			Type:         user.Type,
			RecordStatus: models.RecordStatus{
				IsDeleted: false,
				CreatedAt: dbUser.CreatedAt,
				UpdatedAt: dbUser.UpdatedAt,
				DeletedAt: dbUser.DeletedAt,
			},
		})
	}

	if err := tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot rollback transaction")
		}
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot commit transaction")
	}

	return insertedUsers, nil
}

func (r *sqlxUserRepository) SetPassword(ctx context.Context, ID string, password string) error {
	query := `
	INSERT INTO user_passwords (user_id,password)
	VALUES ($1,$2)
	ON CONFLICT (user_id) DO UPDATE
	SET password = $2
	`

	_, err := r.db.ExecContext(ctx, query, ID, password)
	if err != nil {
		return err
	}

	return nil
}

type updateUser struct {
	ID           string
	Username     string         `db:"username"`
	DisplayName  string         `db:"display_name"`
	Roles        pq.StringArray `db:"roles"`
	Email        *string        `db:"email"`
	ProfileImage *string        `db:"profile_image"`
}

func (r *sqlxUserRepository) Update(ctx context.Context, ID string, user *requests.UpdateUser) (*models.User, error) {
	fields := &updateUser{
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
	RETURNING *
	`, updateFields)

	query, args, err := sqlx.Named(query, fields)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	row, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		log.Println(err)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
			}
		}
		return nil, err
	}

	var updatedUser PostgresUser
	for row.Next() {
		err = row.StructScan(&updatedUser)
		if err != nil {
			return nil, err
		}
	}

	return &models.User{
		ID:           updatedUser.ID,
		Email:        updatedUser.Email,
		Username:     updatedUser.Username,
		DisplayName:  updatedUser.DisplayName,
		ProfileImage: updatedUser.ProfileImage,
		Roles:        updatedUser.Roles,
		Type:         updatedUser.Type,
		RecordStatus: models.RecordStatus{
			IsDeleted: updatedUser.IsDeleted,
			CreatedAt: updatedUser.CreatedAt,
			UpdatedAt: updatedUser.UpdatedAt,
			DeletedAt: updatedUser.DeletedAt,
		},
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
