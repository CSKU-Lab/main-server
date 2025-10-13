package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxSemesterRepository struct {
	db *sqlx.DB
}

type semester struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Type        string    `db:"type"`
	StartedDate time.Time `db:"started_date"`
}

func NewSqlxSemesterRepository(db *sqlx.DB) repositories.SemesterRepository {
	return &sqlxSemesterRepository{
		db: db,
	}
}

func (r *sqlxSemesterRepository) Create(ctx context.Context, ID string, sem *requests.CreateSemester) error {
	query := `INSERT INTO semesters (
		id,
		name,
		type,
		started_date
	) VALUES ($1,$2,$3,$4)`
	_, err := r.db.ExecContext(ctx, query, ID, sem.Name, sem.Type, sem.StartedDate)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Code:       cserrors.SemesterAlreadyExists,
					Message:    "Semester already exists",
				})
			}
		}
		return err
	}

	return nil

}

func (r *sqlxSemesterRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Semester, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT id,name,type,started_date FROM semesters
		WHERE (name ILIKE $1
		OR type::text ILIKE $1
		OR DATE(started_date)::text = $1)
		AND deleted_at IS NULL`

	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+2, len(filterArgs)+3)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	pgSemesters := []semester{}
	err := r.db.SelectContext(ctx, &pgSemesters, query, args...)
	if err != nil {
		return nil, err
	}

	semesters := make([]models.Semester, len(pgSemesters))
	for i, pgSemester := range pgSemesters {
		semesters[i] = models.Semester{
			ID:        pgSemester.ID,
			Name:      pgSemester.Name,
			Type:      models.SemesterType(pgSemester.Type),
			StartDate: pgSemester.StartedDate,
		}
	}

	return semesters, nil
}

func (r *sqlxSemesterRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT COUNT(*) FROM semesters
		WHERE (name LIKE $1
		OR type::text LIKE $1
		OR DATE(started_date)::text = $1) AND
		deleted_at IS NULL`

	query := fmt.Sprintf(`%s%s`, baseQuery, filterWhereClause)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxSemesterRepository) GetByID(ctx context.Context, ID string) (*models.Semester, error) {
	row := r.db.QueryRowxContext(ctx, "SELECT id,name,type,started_date FROM semesters WHERE id = $1 AND is_deleted = false", ID)

	var sem semester

	err := row.StructScan(&sem)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Code:       cserrors.SemesterNotFound,
				Message:    "Semester not found",
			})
		}
		return nil, err
	}

	return &models.Semester{
		ID:        sem.ID,
		Name:      sem.Name,
		Type:      models.SemesterType(sem.Type),
		StartDate: sem.StartedDate,
	}, nil
}

func (r *sqlxSemesterRepository) UpdateByID(ctx context.Context, ID string, req *requests.UpdateSemester) error {
	_, err := r.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	sem := &semester{
		ID:          ID,
		Name:        req.Name,
		StartedDate: req.StartedDate,
		Type:        string(req.Type),
	}

	updateFields := getUpdateFields(sem)

	query := fmt.Sprintf(`
	UPDATE semesters
	SET %s , updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	_, err = r.db.NamedExecContext(ctx, query, sem)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "Semester not found",
				})
			}
		}
		return err
	}

	return nil
}

func (r *sqlxSemesterRepository) DeleteByID(ctx context.Context, ID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE semesters SET is_deleted = true, deleted_at = NOW() WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
