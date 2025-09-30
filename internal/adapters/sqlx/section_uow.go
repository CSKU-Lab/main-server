package sqlx

import (
	"database/sql"
	"log"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

type sectionUowImpl struct {
	db *sqlx.DB
}

type sectionUowInstance struct {
	tx *sqlx.Tx
}

func NewSectionUoWRepository(ctx context.Context, db *sqlx.DB) repositories.SectionUoWRepository {
	return &sectionUowImpl{
		db: db,
	}
}

func (s *sectionUowInstance) Section() repositories.SectionRepository {
	return NewSectionRepository(s.tx)
}

func (s *sectionUowInstance) SectionInstructor() repositories.SectionInstructorRepository {
	return NewSectionInstructorRepository(s.tx)
}

func (s *sectionUowInstance) SectionStudent() repositories.SectionStudentRepository {
	return NewSectionStudentRepository(s.tx)
}

func (s *sectionUowImpl) Execute(ctx context.Context, cb func(s repositories.SectionUoWInstance) error) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatalln("Cannot start transaction in sqlx SectionUowRepository")
	}
	defer tx.Rollback()

	uow := &sectionUowInstance{tx: tx}
	err = cb(uow)
	if err != nil {
		return err
	}

	return tx.Commit()
}
