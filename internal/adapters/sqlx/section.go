package sqlx

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxSectionRepository struct {
	db instance
}

type sectionSchema struct {
	ID     string  `db:"id"`
	Name   string  `db:"name"`
	Banner *string `db:"banner"`
}

type rawSectionSchema struct {
	ID         string  `db:"id"`
	Name       string  `db:"name"`
	Banner     *string `db:"banner"`
	CourseID   string  `db:"course_id"`
	SemesterID string  `db:"semester_id"`
	CreatedAt  string  `db:"created_at"`
	UpdatedAt  string  `db:"updated_at"`
}

func NewSectionRepository(db instance) repositories.SectionRepository {
	return &sqlxSectionRepository{db: db}
}

func (s *sqlxSectionRepository) Create(ctx context.Context, ID string, section *repositories.CreateSection) error {
	query := "INSERT INTO sections (id, name, course_id, semester_id) VALUES ($1, $2, $3, $4)"

	_, err := s.db.ExecContext(ctx, query, ID, section.Name, section.CourseID, section.SemesterID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusConflict, Message: "Section already exists"})
			}

			if pqErr.Code.Name() == "foreign_key_violation" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "Course ID or Section ID does not exists",
				})
			}
		}
		return err
	}

	return nil
}

func (s *sqlxSectionRepository) UpdateByID(ctx context.Context, ID string, section *repositories.UpdateSection) error {
	updatedSchema := &sectionSchema{
		ID:     ID,
		Name:   section.Name,
		Banner: section.Banner,
	}

	updateFields := getUpdateFields(updatedSchema)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE sections
	SET %s , updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	query, args, err := sqlx.Named(query, updatedSchema)
	if err != nil {
		return err
	}

	query = s.db.Rebind(query)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlxSectionRepository) GetByID(ctx context.Context, ID string) (*models.Section, error) {
	var section sectionSchema
	query := "SELECT id, name, banner FROM sections WHERE id = $1 AND is_deleted = false"
	err := s.db.GetContext(ctx, &section, query, ID)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Section not found"})
	}

	return &models.Section{
		ID:     section.ID,
		Name:   section.Name,
		Banner: section.Banner,
	}, nil
}

func (s *sqlxSectionRepository) GetBySemesterID(ctx context.Context, ID string) ([]models.Section, error) {
	var dbSections []sectionSchema
	query := "SELECT id, name, banner FROM sections WHERE semester_id = $1 AND is_deleted = false"
	err := s.db.SelectContext(ctx, &dbSections, query, ID)
	if err != nil {
		return nil, err
	}

	var sections []models.Section
	for _, dbSection := range dbSections {
		sections = append(sections, models.Section{
			ID:     dbSection.ID,
			Name:   dbSection.Name,
			Banner: dbSection.Banner,
		})
	}

	return sections, nil
}

func (s *sqlxSectionRepository) GetRawBySemesterID(ctx context.Context, ID string) ([]repositories.RawSection, error) {
	var dbSections []rawSectionSchema
	query := "SELECT id, name, banner, course_id, semester_id, created_at, updated_at FROM sections WHERE semester_id = $1 AND is_deleted = false"
	err := s.db.SelectContext(ctx, &dbSections, query, ID)
	if err != nil {
		return nil, err
	}

	var sections []repositories.RawSection
	for _, dbSection := range dbSections {
		sections = append(sections, repositories.RawSection{
			ID:         dbSection.ID,
			Name:       dbSection.Name,
			Banner:     dbSection.Banner,
			CourseID:   dbSection.CourseID,
			SemesterID: dbSection.SemesterID,
			CreatedAt:  dbSection.CreatedAt,
			UpdatedAt:  dbSection.UpdatedAt,
		})
	}

	return sections, nil
}

func (s *sqlxSectionRepository) DeleteByID(ctx context.Context, ID string) error {
	query := "UPDATE sections SET is_deleted = true, deleted_at = NOW() WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, ID)
	if err != nil {
		return err
	}
	return nil
}
