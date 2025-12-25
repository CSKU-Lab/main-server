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

type defaultLabSchema struct {
	ID        string    `db:"id"`
	LabID     string    `db:"lab_id"`
	LabName   string    `db:"lab_name"`
	CourseID  string    `db:"course_id"`
	Position  int       `db:"position"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sqlxDefaultLabRepository struct {
	db instance
}

func NewSqlxDefaultLabRepository(db instance) repositories.DefaultLabRepository {
	return &sqlxDefaultLabRepository{
		db: db,
	}
}

func (dl *sqlxDefaultLabRepository) Create(ctx context.Context, req *requests.SetDefaultLab, id string, courseID string, labName string) error {
	query := `INSERT INTO default_labs (lab_id, course_id, id, position, lab_name) VALUES ($1, $2, $3, $4, $5)`
	_, err := dl.db.ExecContext(ctx, query, req.LabID, courseID, id, req.Position, labName)
	if err != nil {
		return err
	}

	return nil
}

func (dl *sqlxDefaultLabRepository) Update(ctx context.Context, req *requests.UpdateDefaultLab, id string) error {
	updatedSchema := &defaultLabSchema{
		ID:       id,
		LabID:    req.LabID,
		Position: req.Position,
	}

	updateFields := getUpdateFields(updatedSchema)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE default_labs
	SET %s , updated_at = NOW()
	WHERE id = :id
		AND is_deleted = false
	`, updateFields)

	query, args, err := sqlx.Named(query, updatedSchema)
	if err != nil {
		return err
	}

	query = dl.db.Rebind(query)

	_, err = dl.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (dl *sqlxDefaultLabRepository) GetByID(ctx context.Context, labID string, courseID string) (*models.DefaultLab, error) {
	query := `SELECT id, lab_id, course_id, position, created_at, updated_at FROM default_labs WHERE lab_id = $1 AND course_id = $2 AND is_deleted = false`

	defaultLabSchema := &defaultLabSchema{}
	err := dl.db.GetContext(ctx, defaultLabSchema, query, labID, courseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "DefaultLab not found"})
		}
		return nil, err
	}

	return &models.DefaultLab{
		ID:        defaultLabSchema.ID,
		LabID:     defaultLabSchema.LabID,
		CourseID:  defaultLabSchema.CourseID,
		Position:  defaultLabSchema.Position,
		CreatedAt: defaultLabSchema.CreatedAt,
		UpdatedAt: defaultLabSchema.UpdatedAt,
	}, nil
}

func (dl *sqlxDefaultLabRepository) DeleteByID(ctx context.Context, id string) error {
	query := "UPDATE default_labs SET is_deleted = true, deleted_at = NOW() WHERE id = $1"
	_, err := dl.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (dl *sqlxDefaultLabRepository) GetPagination(
	ctx context.Context,
	page int,
	limit int,
	search string,
	sortBy string,
	sortOrder string,
	filters []sanitize.Filter,
) ([]models.DefaultLab, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)
	baseQuery := `
		SELECT id, lab_id, lab_name, course_id, position, created_at, updated_at
		FROM default_labs
		WHERE is_deleted = false
		  AND lab_name ILIKE $1
	`

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)

	var query string

	if limit == -1 {
		query = fmt.Sprintf(`
			%s%s
			ORDER BY %s %s
		`,
			baseQuery,
			filterWhereClause,
			sortBy,
			sortOrder,
		)
	} else {
		query = fmt.Sprintf(`
			%s%s
			ORDER BY %s %s
			OFFSET $%d
			LIMIT $%d
		`,
			baseQuery,
			filterWhereClause,
			sortBy,
			sortOrder,
			len(args)+1,
			len(args)+2,
		)

		args = append(args, (page-1)*limit, limit)
	}

	defaultLabsSchema := []defaultLabSchema{}
	if err := dl.db.SelectContext(ctx, &defaultLabsSchema, query, args...); err != nil {
		return nil, err
	}

	defaultLabs := make([]models.DefaultLab, 0, len(defaultLabsSchema))
	for _, dl := range defaultLabsSchema {
		defaultLabs = append(defaultLabs, models.DefaultLab{
			ID:        dl.ID,
			LabID:     dl.LabID,
			LabName:   dl.LabName,
			CourseID:  dl.CourseID,
			Position:  dl.Position,
			CreatedAt: dl.CreatedAt,
			UpdatedAt: dl.UpdatedAt,
		})
	}

	return defaultLabs, nil
}

func (dl *sqlxDefaultLabRepository) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT COUNT(*) FROM default_labs WHERE is_deleted = false`

	query := baseQuery + filterWhereClause
	var count int
	err := dl.db.GetContext(ctx, &count, query, filterArgs...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dl *sqlxDefaultLabRepository) ShiftInsertedPositions(
	ctx context.Context,
	courseID string,
	currPos int,
	reqPos int,
) error {
	if currPos == reqPos {
		return nil
	}

	if currPos > reqPos {
		_, err := dl.db.ExecContext(ctx, `
			UPDATE default_labs
			SET position = position + 1
			WHERE course_id = $1
			  AND position >= $2
			  AND position < $3
			  AND is_deleted = false
		`, courseID, reqPos, currPos)
		return err
	}

	_, err := dl.db.ExecContext(ctx, `
		UPDATE default_labs
		SET position = position - 1
		WHERE course_id = $1
		  AND position > $2
		  AND position <= $3
		  AND is_deleted = false
	`, courseID, currPos, reqPos)

	return err
}

func (dl *sqlxDefaultLabRepository) ShiftDownPositions(ctx context.Context, courseID string, position int) error {
	_, err := dl.db.ExecContext(ctx, `
		UPDATE default_labs
		SET position = position + 1
		WHERE course_id = $1
		  AND position >= $2 
			AND is_deleted = false
	`, courseID, position)
	if err != nil {
		return err
	}
	return nil
}

func (dl *sqlxDefaultLabRepository) ShiftUpPositions(ctx context.Context, courseID string, labID string, position int) error {
	_, err := dl.db.ExecContext(ctx, `
		UPDATE default_labs
		SET position = position - 1
		WHERE course_id = $1
		  AND position >= $2
			AND is_deleted = false
			AND lab_id != $3
	`, courseID, position, labID)
	if err != nil {
		return err
	}
	return nil
}

func (dl *sqlxDefaultLabRepository) GetMaxPosition(ctx context.Context, courseID string, labID string) (int, error) {
	var max int

	err := dl.db.QueryRowxContext(ctx, `
		SELECT COALESCE(MAX(position), 0)
		FROM default_labs
		WHERE course_id = $1 
			AND is_deleted = false
			AND lab_id != $2
	`, courseID, labID).Scan(&max)

	return max + 1, err
}

func (dl *sqlxDefaultLabRepository) GetByCourseID(
	ctx context.Context,
	courseID string,
) ([]models.DefaultLab, error) {
	query := `
		SELECT id, lab_id, course_id, position, created_at, updated_at
		FROM default_labs
		WHERE course_id = $1 AND is_deleted = false
	`

	defaultLabsSchema := []defaultLabSchema{}

	err := dl.db.SelectContext(ctx, &defaultLabsSchema, query, courseID)
	if err != nil {
		return nil, err
	}

	defaultLabs := make([]models.DefaultLab, 0, len(defaultLabsSchema))
	for _, ls := range defaultLabsSchema {
		defaultLabs = append(defaultLabs, models.DefaultLab{
			ID:        ls.ID,
			LabID:     ls.LabID,
			CourseID:  ls.CourseID,
			Position:  ls.Position,
			CreatedAt: ls.CreatedAt,
			UpdatedAt: ls.UpdatedAt,
		})
	}

	return defaultLabs, nil
}

func (dl *sqlxDefaultLabRepository) GetByLabID(
	ctx context.Context,
	labID string,
) ([]models.DefaultLab, error) {
	query := `
		SELECT id, lab_id, course_id, position, created_at, updated_at
		FROM default_labs
		WHERE lab_id = $1 AND is_deleted = false
	`

	defaultLabsSchema := []defaultLabSchema{}

	err := dl.db.SelectContext(ctx, &defaultLabsSchema, query, labID)
	if err != nil {
		return nil, err
	}

	defaultLabs := make([]models.DefaultLab, 0, len(defaultLabsSchema))
	for _, ls := range defaultLabsSchema {
		defaultLabs = append(defaultLabs, models.DefaultLab{
			ID:        ls.ID,
			LabID:     ls.LabID,
			CourseID:  ls.CourseID,
			Position:  ls.Position,
			CreatedAt: ls.CreatedAt,
			UpdatedAt: ls.UpdatedAt,
		})
	}

	return defaultLabs, nil
}
