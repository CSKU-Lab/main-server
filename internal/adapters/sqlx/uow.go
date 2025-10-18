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

func NewUoWRepository(ctx context.Context, db *sqlx.DB) repositories.UoWRepository {
	return &uowImpl{db: db,
		ctx: ctx,
	}
}

func (u *uowInstance) User() repositories.User {
	return NewUserRepository(u.tx)
}

func (u *uowInstance) UserPassword() repositories.UserPassword {
	return NewUserPasswordRepository(u.tx)
}

func (u *uowInstance) UserGroup() repositories.UserGroup {
	return NewUserGroupRepository(u.tx)
}

func (u *uowInstance) Section() repositories.SectionRepository {
	return NewSectionRepository(u.tx)
}

func (u *uowInstance) SectionInstructor() repositories.SectionInstructorRepository {
	return NewSectionInstructorRepository(u.tx)
}

func (u *uowInstance) SectionStudent() repositories.SectionStudentRepository {
	return NewSectionStudentRepository(u.tx)
}

func (s *uowImpl) Execute(ctx context.Context, cb func(s repositories.UoWInstance) error) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatalln("Cannot start transaction in sqlx UowRepository")
	}
	defer tx.Rollback()

	uow := &uowInstance{tx: tx}
	err = cb(uow)
	if err != nil {
		return err
	}

	return tx.Commit()
}
