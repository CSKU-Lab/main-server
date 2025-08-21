package sqlx

import (
	"database/sql"
	"log"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

type uowImpl struct {
	db  *sqlx.DB
	ctx context.Context
}

type uowInstance struct {
	tx *sqlx.Tx
}

func NewUserUoWRepository(ctx context.Context, db *sqlx.DB) repositories.UserUoWRepository {
	return &uowImpl{db: db,
		ctx: ctx,
	}
}

func (u *uowInstance) User() repositories.User {
	return NewUserRepository(u.tx)
}

func (u *uowInstance) UserPassword() repositories.UserPasswordRepository {
	return NewUserPasswordRepository(u.tx)
}

func (u *uowInstance) UserGroup() repositories.UserGroupRepository {
	return NewUserGroupRepository(u.tx)
}

func (u *uowImpl) Execute(cb func(u repositories.UserUoWInstance) error) error {
	tx, err := u.db.BeginTxx(u.ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatalln("Cannot start transaction in sqlx UowRepository")
	}
	defer tx.Rollback()

	uow := &uowInstance{tx: tx}
	err = cb(uow)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
