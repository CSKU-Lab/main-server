package sqlx

import (
	"context"
	"database/sql"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
)

type userPasswordRepository struct {
	db instance
}

func NewUserPasswordRepository(db instance) repositories.UserPasswordRepository {
	return &userPasswordRepository{db: db}
}

func (u *userPasswordRepository) GetPasswordByID(ctx context.Context, ID string) (string, error) {
	var password string
	err := u.db.GetContext(ctx, &password, "SELECT password FROM user_passwords WHERE user_id = $1", ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "User not found")
		}
		return "", err
	}

	return password, nil
}

func (u *userPasswordRepository) SetPassword(ctx context.Context, ID string, password string) error {
	query := `
	INSERT INTO user_passwords (user_id,password)
	VALUES ($1,$2)
	ON CONFLICT (user_id) DO UPDATE
	SET password = $2
	`

	_, err := u.db.ExecContext(ctx, query, ID, password)
	if err != nil {
		return err
	}

	return nil
}
