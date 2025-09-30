package sqlx

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type sqlxSectionInstructorRepository struct {
	db instance
}

type sqlxSectionInstructor struct {
	ID           string  `db:"id"`
	Username     string  `db:"username"`
	DisplayName  string  `db:"display_name"`
	ProfileImage *string `db:"profile_image"`
}

func NewSectionInstructorRepository(db instance) repositories.SectionInstructorRepository {
	return &sqlxSectionInstructorRepository{
		db: db,
	}
}

func (s *sqlxSectionInstructorRepository) Add(ctx context.Context, sectionID string, instructorID string) error {
	query := `INSERT INTO section_instructors (section_id,instructor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, sectionID, instructorID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "foreign_key_violation" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: 400,
					Message:    "Section or Instructor not found",
				})
			}
		}
		return err
	}

	return nil
}

func (s *sqlxSectionInstructorRepository) Get(ctx context.Context, sectionID string) ([]models.SectionInstructor, error) {
	query := `SELECT id, username, display_name, profile_image FROM users
		  JOIN section_instructors ON users.id = section_instructors.instructor_id
		  WHERE section_instructors.section_id = $1`

	var instructors []sqlxSectionInstructor
	err := s.db.SelectContext(ctx, &instructors, query, sectionID)
	if err != nil {
		return nil, err
	}

	instructorsModel := make([]models.SectionInstructor, len(instructors))
	for i := range instructors {
		instructorsModel[i] = models.SectionInstructor{
			ID:           instructors[i].ID,
			Username:     instructors[i].Username,
			DisplayName:  instructors[i].DisplayName,
			ProfileImage: instructors[i].ProfileImage,
		}
	}

	return instructorsModel, nil
}

func (s *sqlxSectionInstructorRepository) DeleteBySectionID(ctx context.Context, sectionID string) error {
	query := `DELETE FROM section_instructors WHERE section_id = $1`

	_, err := s.db.ExecContext(ctx, query, sectionID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "foreign_key_violation" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: 400,
					Message:    "Section or Instructor not found",
				})
			}
		}
	}

	return nil
}
