package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/jmoiron/sqlx"
)

type labSectionSchema struct {
	ID        string    `db:"id"`
	LabID     string    `db:"lab_id"`
	SectionID string    `db:"section_id"`
	Position  int       `db:"position"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sqlxLabSectionRepository struct {
	db instance
}

func NewSqlxLabSectionRepository(db instance) repositories.LabSectionRepository {
	return &sqlxLabSectionRepository{db: db}
}

func (ls *sqlxLabSectionRepository) Create(ctx context.Context, req *requests.SetLabSection, id string) error {
	query := `INSERT INTO lab_sections (lab_id, section_id, position, id) VALUES ($1, $2, $3, $4)`
	_, err := ls.db.ExecContext(ctx, query, req.LabID, req.SectionID, req.Position, id)
	if err != nil {
		return err
	}

	return nil
}

func (ls *sqlxLabSectionRepository) ShiftDownPositions(ctx context.Context, sectionID string, position int) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE lab_sections
		SET position = position + 1
		WHERE section_id = $1
		  AND position >= $2
			AND is_deleted = false
	`, sectionID, position)
	if err != nil {
		return err
	}
	return nil
}

func (ls *sqlxLabSectionRepository) ShiftUpPositions(ctx context.Context, sectionID string, position int) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE lab_sections
		SET position = position - 1
		WHERE section_id = $1
		  AND position >= $2
			AND is_deleted = false
	`, sectionID, position)
	if err != nil {
		return err
	}
	return nil
}

func (ls *sqlxLabSectionRepository) GetMaxPosition(ctx context.Context, sectionID string, labID string) (int, error) {
	var max int

	err := ls.db.QueryRowxContext(ctx, `
		SELECT COALESCE(MAX(position), 0)
		FROM lab_sections
		WHERE section_id = $1 
			AND is_deleted = false
			AND lab_id != $2
	`, sectionID, labID).Scan(&max)

	return max + 1, err
}

func (ls *sqlxLabSectionRepository) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.LabSection, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT id, lab_id, section_id, position, created_at, updated_at FROM lab_sections WHERE is_deleted = false`
	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+1, len(filterArgs)+2)

	args := make([]any, 0, len(filterArgs)+2)
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	labSectionsSchema := []labSectionSchema{}
	err := ls.db.SelectContext(ctx, &labSectionsSchema, query, args...)
	if err != nil {
		return nil, err
	}
	labSections := make([]models.LabSection, 0, len(labSectionsSchema))
	for _, labSection := range labSectionsSchema {
		labSections = append(labSections, models.LabSection{
			ID:        labSection.ID,
			LabID:     labSection.LabID,
			SectionID: labSection.SectionID,
			Position:  labSection.Position,
			CreatedAt: labSection.CreatedAt,
			UpdatedAt: labSection.UpdatedAt,
		})
	}
	return labSections, nil
}

func (ls *sqlxLabSectionRepository) UpdateByID(ctx context.Context, labID string, sectionID string, id string, req *requests.UpdateLabSection) error {
	updatedSchema := &labSectionSchema{
		ID:        id,
		LabID:     labID,
		SectionID: sectionID,
		Position:  req.Position,
	}

	updateFields := getUpdateFields(updatedSchema)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE lab_sections
	SET %s , updated_at = NOW()
		WHERE lab_id = :lab_id AND section_id = :section_id AND id = :id
		AND is_deleted = false
	`, updateFields)

	query, args, err := sqlx.Named(query, updatedSchema)
	if err != nil {
		return err
	}

	query = ls.db.Rebind(query)

	_, err = ls.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (ls *sqlxLabSectionRepository) DeleteByID(ctx context.Context, id string) error {
	query := "UPDATE lab_sections SET is_deleted = true, deleted_at = NOW() WHERE id = $1"
	_, err := ls.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (ls *sqlxLabSectionRepository) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT COUNT(*) FROM lab_sections WHERE is_deleted = false`

	query := baseQuery + filterWhereClause
	var count int
	err := ls.db.GetContext(ctx, &count, query, filterArgs...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ls *sqlxLabSectionRepository) GetByID(ctx context.Context, labID string, sectionID string) (*models.LabSection, error) {
	query := `SELECT id, lab_id, section_id, position, created_at, updated_at FROM lab_sections WHERE lab_id = $1 AND section_id = $2 AND is_deleted = false`

	labSectionSchema := &labSectionSchema{}
	err := ls.db.GetContext(ctx, labSectionSchema, query, labID, sectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "LabSection not found"})
		}
		return nil, err
	}

	return &models.LabSection{
		ID:        labSectionSchema.ID,
		LabID:     labSectionSchema.LabID,
		SectionID: labSectionSchema.SectionID,
		Position:  labSectionSchema.Position,
		CreatedAt: labSectionSchema.CreatedAt,
		UpdatedAt: labSectionSchema.UpdatedAt,
	}, nil
}
