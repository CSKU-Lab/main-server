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
)

type defaultLabSchema struct {
	ID        string    `db:"id"`
	LabID     string    `db:"lab_id"`
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

func (dl *sqlxDefaultLabRepository) Create(ctx context.Context, req *requests.SetDefaultLab, id string, courseID string) error {
	query := `INSERT INTO default_labs (lab_id, course_id, id, position) VALUES ($1, $2, $3, $4)`
	_, err := dl.db.ExecContext(ctx, query, req.LabID, courseID, id, req.Position)
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

func (dl *sqlxDefaultLabRepository) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.DefaultLab, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT id, lab_id, course_id, position, created_at, updated_at FROM default_labs WHERE is_deleted = false`
	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+1, len(filterArgs)+2)

	args := make([]any, 0, len(filterArgs)+2)
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	defaultLabsSchema := []defaultLabSchema{}
	err := dl.db.SelectContext(ctx, &defaultLabsSchema, query, args...)
	if err != nil {
		return nil, err
	}
	defaultLabs := make([]models.DefaultLab, 0, len(defaultLabsSchema))
	for _, defaultLab := range defaultLabsSchema {
		defaultLabs = append(defaultLabs, models.DefaultLab{
			ID:        defaultLab.ID,
			LabID:     defaultLab.LabID,
			CourseID:  defaultLab.CourseID,
			Position:  defaultLab.Position,
			CreatedAt: defaultLab.CreatedAt,
			UpdatedAt: defaultLab.UpdatedAt,
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
	labID string,
) ([]models.DefaultLab, error) {
	query := `
		SELECT id, lab_id, course_id, position, created_at, updated_at
		FROM default_labs
		WHERE course_id = $1 AND is_deleted = false
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
