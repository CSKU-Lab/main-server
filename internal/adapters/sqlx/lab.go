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

type labSchema struct {
	ID          string    `db:"id"`
	DisplayName string    `db:"display_name"`
	CourseID    string    `db:"course_id"`
	CreatedBy   string    `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type sqlxLabRepository struct {
	db instance
}

func NewSqlxLabRepository(db instance) repositories.LabRepository {
	return &sqlxLabRepository{
		db: db,
	}
}

func (l *sqlxLabRepository) GetByID(ctx context.Context, labID string) (*models.Lab, error) {
	query := `SELECT id, display_name, course_id, created_by, created_at, updated_at FROM labs WHERE id = $1`

	lab := &labSchema{}
	err := l.db.GetContext(ctx, lab, query, labID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Lab not found"})
		}
		return nil, err
	}

	return &models.Lab{
		ID:          lab.ID,
		DisplayName: lab.DisplayName,
		CourseID:    lab.CourseID,
		CreatedBy:   lab.CreatedBy,
		CreatedAt:   lab.CreatedAt,
		UpdatedAt:   lab.UpdatedAt,
	}, nil
}

func (l *sqlxLabRepository) Create(ctx context.Context, id string, req *requests.CreateLab, userID string) error {
	query := `INSERT INTO labs (id, display_name, course_id, created_by) VALUES ($1, $2, $3, $4)`
	_, err := l.db.ExecContext(ctx, query, id, req.DisplayName, req.CourseID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (l *sqlxLabRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Lab, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT id, display_name, course_id, created_by, created_at, updated_at FROM labs WHERE (display_name ILIKE $1)`
	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+2, len(filterArgs)+3)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	labsSchema := []labSchema{}
	err := l.db.SelectContext(ctx, &labsSchema, query, args...)
	if err != nil {
		return nil, err
	}
	labs := make([]models.Lab, 0, len(labsSchema))
	for _, s := range labsSchema {
		labs = append(labs, models.Lab{
			ID:          s.ID,
			DisplayName: s.DisplayName,
			CourseID:    s.CourseID,
			CreatedBy:   s.CreatedBy,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}
	return labs, nil
}

func (l *sqlxLabRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT COUNT(*) FROM labs
		WHERE (display_name ILIKE $1)`

	query := fmt.Sprintf(`%s%s`, baseQuery, filterWhereClause)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)

	var count int
	err := l.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *sqlxLabRepository) UpdateByID(ctx context.Context, labID string, req *requests.BaseUpdateLab) error {
	updatedSchema := &labSchema{
		ID:          labID,
		DisplayName: req.DisplayName,
		CourseID:    req.CourseID,
	}

	updateFields := getUpdateFields(updatedSchema)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE labs
	SET %s , updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	query, args, err := sqlx.Named(query, updatedSchema)
	if err != nil {
		return err
	}

	query = l.db.Rebind(query)

	_, err = l.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (l *sqlxLabRepository) DeleteByID(ctx context.Context, labID string) error {
	query := "UPDATE labs SET is_deleted = true, deleted_at = NOW() WHERE id = $1"
	_, err := l.db.ExecContext(ctx, query, labID)
	if err != nil {
		return err
	}
	return nil
}
