package sqlx

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type sectionStudentRepository struct {
	db instance
}

type student struct {
	ID           string  `db:"id"`
	Username     string  `db:"username"`
	DisplayName  string  `db:"display_name"`
	ProfileImage *string `db:"profile_image"`
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
			if pqErr.Code.Name() == "unique_violation" {
				return cserrors.New(&cserrors.Option{
					Code:       cserrors.UniqueViolation,
					HttpStatus: 400,
					Message:    "Student already added to section",
				})
			}
		}
		return err
	}
	return nil
}

func (s *sectionStudentRepository) RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error {
	query := `DELETE FROM section_students WHERE section_id = $1 AND student_id = $2`
	_, err := s.db.ExecContext(ctx, query, sectionID, studentID)
	if err != nil {
		return err
	}
	return nil
}

func (s *sectionStudentRepository) GetBySectionID(ctx context.Context, sectionID string) ([]models.Student, error) {
	query := `SELECT id, username, display_name, profile_image FROM section_students ss
		  JOIN users u ON ss.student_id = u.id
		  WHERE ss.section_id = $1`

	dbStudents := []student{}
	err := s.db.SelectContext(ctx, &dbStudents, query, sectionID)
	if err != nil {
		return nil, err
	}

	students := make([]models.Student, 0, len(dbStudents))
	for _, dbStudent := range dbStudents {
		students = append(students, models.Student{
			ID:           dbStudent.ID,
			Username:     dbStudent.Username,
			DisplayName:  dbStudent.DisplayName,
			ProfileImage: dbStudent.ProfileImage,
		})
	}

	return students, nil
}
