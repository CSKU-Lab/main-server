package sqlx

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type sectionStudentRepository struct {
	db instance
}

func NewSectionStudentRepository(db instance) repositories.SectionStudentRepository {
	return &sectionStudentRepository{db: db}
}

func (s *sectionStudentRepository) Add(ctx context.Context, sectionID string, studentID string) error {
	query := `INSERT INTO section_students (section_id, student_id) VALUES ($1, $2)`
	_, err := s.db.ExecContext(ctx, query, sectionID, studentID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "foreign_key_violation" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: 400,
					Message:    "Section or Student not found",
				})
			}
		}
		return err
	}
	return nil
}

func (s *sectionStudentRepository) DeleteBySectionID(ctx context.Context, sectionID string) error {
	query := `DELETE FROM section_students WHERE section_id = $1`
	_, err := s.db.ExecContext(ctx, query, sectionID)
	if err != nil {
		return err
	}
	return nil
}
