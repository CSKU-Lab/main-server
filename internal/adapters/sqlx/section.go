package sqlx

import (
	"context"
	"errors"
	"fmt"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxSectionRepository struct {
	db *sqlx.DB
}

type sectionSchema struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	Image     *string `db:"image"`
	StartedAt string  `db:"started_at"`
	EndedAt   string  `db:"ended_at"`
}

func NewSectionRepository(db *sqlx.DB) repositories.SectionRepository {
	return &sqlxSectionRepository{db: db}
}

func (s *sqlxSectionRepository) Create(ctx context.Context, section *models.Section, courseID, semesterID string) error {
	query := "INSERT INTO sections (id, name, image, started_at, ended_at, course_id, semester_id) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	_, err := s.db.ExecContext(ctx, query, section.ID, section.Name, section.Image, section.StartedAt, section.EndedAt, courseID, semesterID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(cserrors.ALREADY_EXISTS, "Section already exists")
			}
		}
		return err
	}

	return nil
}

func (s *sqlxSectionRepository) UpdateByID(ctx context.Context, section *models.Section) error {
	updatedSchema := &sectionSchema{
		ID:        section.ID,
		Name:      section.Name,
		Image:     section.Image,
		StartedAt: section.StartedAt,
		EndedAt:   section.EndedAt,
	}

	updateFields := getUpdateFields(updatedSchema)

	query := fmt.Sprintf(`
	UPDATE sections
	SET %s , updated_at = NOW()
	WHERE id = :id
	RETURNING *
	`, updateFields)

	_, err := s.db.NamedExecContext(ctx, query, updatedSchema)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlxSectionRepository) GetByID(ctx context.Context, ID string) (*models.Section, error) {
	var section sectionSchema
	query := "SELECT id, name, image, started_at, ended_at FROM sections WHERE id = $1 AND is_deleted = false"
	err := s.db.GetContext(ctx, &section, query, ID)
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Section not found")
	}

	return &models.Section{
		ID:        section.ID,
		Name:      section.Name,
		Image:     section.Image,
		StartedAt: section.StartedAt,
		EndedAt:   section.EndedAt,
	}, nil
}

func (s *sqlxSectionRepository) GetBySemesterID(ctx context.Context, ID string) ([]models.Section, error) {
	var sections []models.Section
	query := "SELECT * FROM sections WHERE section_id = $1 AND is_deleted = false"
	err := s.db.SelectContext(ctx, &sections, query, ID)
	if err != nil {
		return nil, err
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
