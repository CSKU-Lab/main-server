package repositories

import "context"

type UserPasswordRepository interface {
	GetPasswordByID(ctx context.Context, ID string) (string, error)
	SetPassword(ctx context.Context, username string, password string) error
}
