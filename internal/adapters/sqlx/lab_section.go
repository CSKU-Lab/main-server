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
	ID        string     `db:"id"`
	LabID     string     `db:"lab_id"`
	SectionID string     `db:"section_id"`
	Position  int        `db:"position"`
	Status    string     `db:"status"`
	OpenedAt  *time.Time `db:"opened_at"`
	ClosedAt  *time.Time `db:"closed_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type sqlxLabSectionRepository struct {
	db instance
}

func NewSqlxLabSectionRepository(db instance) repositories.LabSectionRepository {
	return &sqlxLabSectionRepository{db: db}
}

func (ls *sqlxLabSectionRepository) Create(ctx context.Context, params repositories.CreateLabSectionParams) error {
	query := `INSERT INTO lab_sections (lab_id, section_id, position, id, status, opened_at, closed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := ls.db.ExecContext(ctx, query, params.LabID, params.SectionID, params.Position, params.ID, params.Status, params.OpenedAt, params.ClosedAt)
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

func (ls *sqlxLabSectionRepository) ShiftRangeDown(ctx context.Context, sectionID string, startPos int, endPos int) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE lab_sections
		SET position = position + 1
		WHERE section_id = $1
		  AND position >= $2
		  AND position <= $3
			AND is_deleted = false
	`, sectionID, startPos, endPos)
	if err != nil {
		return err
	}
	return nil
}

func (ls *sqlxLabSectionRepository) ShiftRangeUp(ctx context.Context, sectionID string, startPos int, endPos int) error {
	_, err := ls.db.ExecContext(ctx, `
		UPDATE lab_sections
		SET position = position - 1
		WHERE section_id = $1
		  AND position >= $2
		  AND position <= $3
			AND is_deleted = false
	`, sectionID, startPos, endPos)
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
	`, sectionID).Scan(&max)

	return max + 1, err
}

func (ls *sqlxLabSectionRepository) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.LabSection, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT ls.id, ls.lab_id, ls.section_id, ls.position, ls.status, ls.opened_at, ls.closed_at, ls.created_at, ls.updated_at
		FROM lab_sections ls
		JOIN labs l ON ls.lab_id = l.id
		WHERE ls.is_deleted = false`
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
			Status:    labSection.Status,
			OpenedAt:  labSection.OpenedAt,
			ClosedAt:  labSection.ClosedAt,
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

func (ls *sqlxLabSectionRepository) UpdateStatusByID(ctx context.Context, labID string, sectionID string, id string, req *requests.UpdateLabSectionStatus) error {
	updatedSchema := &labSectionSchema{
		ID:        id,
		LabID:     labID,
		SectionID: sectionID,
	}

	if req.Status != nil {
		updatedSchema.Status = *req.Status
	}
	if req.OpenedAt != nil {
		updatedSchema.OpenedAt = req.OpenedAt
	}
	if req.ClosedAt != nil {
		updatedSchema.ClosedAt = req.ClosedAt
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

	query =
		ls.db.Rebind(query)

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

	baseQuery := `SELECT COUNT(*) FROM lab_sections ls JOIN labs l ON ls.lab_id = l.id WHERE ls.is_deleted = false`

	query := baseQuery + filterWhereClause
	var count int
	err := ls.db.GetContext(ctx, &count, query, filterArgs...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ls *sqlxLabSectionRepository) GetBySectionID(ctx context.Context, sectionID string) ([]models.Lab, error) {
	query := `SELECT l.id, l.display_name, l.course_id FROM lab_sections ls
		  JOIN labs l ON ls.lab_id = l.id
		  WHERE ls.section_id = $1
		    AND ls.is_deleted = false
		    AND l.is_deleted = false
		    AND ls.status IN ('open', 'readonly', 'disabled')
		  ORDER BY ls.position ASC`

	dbLabs := []labSchema{}
	err := ls.db.SelectContext(ctx, &dbLabs, query, sectionID)
	if err != nil {
		return nil, err
	}

	labs := make([]models.Lab, 0, len(dbLabs))
	for _, dbLab := range dbLabs {
		labs = append(labs, models.Lab{
			ID:          dbLab.ID,
			DisplayName: dbLab.DisplayName,
			CourseID:    dbLab.CourseID,
		})
	}

	return labs, nil
}

func (ls *sqlxLabSectionRepository) GetByID(ctx context.Context, labID string, sectionID string) (*models.LabSection, error) {
	query := `SELECT id, lab_id, section_id, position, status, opened_at, closed_at, created_at, updated_at FROM lab_sections WHERE lab_id = $1 AND section_id = $2 AND is_deleted = false`

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
		Status:    labSectionSchema.Status,
		OpenedAt:  labSectionSchema.OpenedAt,
		ClosedAt:  labSectionSchema.ClosedAt,
		CreatedAt: labSectionSchema.CreatedAt,
		UpdatedAt: labSectionSchema.UpdatedAt,
	}, nil
}

func (ls *sqlxLabSectionRepository) GetByLabID(
	ctx context.Context,
	labID string,
) ([]models.LabSection, error) {
	query := `
		SELECT id, lab_id, section_id, position, status, opened_at, closed_at, created_at, updated_at
		FROM lab_sections
		WHERE lab_id = $1 AND is_deleted = false
	`

	labSectionsSchema := []labSectionSchema{}

	err := ls.db.SelectContext(ctx, &labSectionsSchema, query, labID)
	if err != nil {
		return nil, err
	}

	labSections := make([]models.LabSection, 0, len(labSectionsSchema))
	for _, ls := range labSectionsSchema {
		labSections = append(labSections, models.LabSection{
			ID:        ls.ID,
			LabID:     ls.LabID,
			SectionID: ls.SectionID,
			Position:  ls.Position,
			Status:    ls.Status,
			OpenedAt:  ls.OpenedAt,
			ClosedAt:  ls.ClosedAt,
			CreatedAt: ls.CreatedAt,
			UpdatedAt: ls.UpdatedAt,
		})
	}

	return labSections, nil
}
