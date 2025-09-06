package sqlx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
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
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Semester already exists"})
			}
		}
		return err
	}

	return nil

}

func (r *sqlxSemesterRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error) {
	query := fmt.Sprintf(`SELECT id,name,type,started_date FROM semesters 
		WHERE (name ILIKE $1 
		OR type::text ILIKE $1
		OR DATE(started_date)::text = $1)
		AND deleted_at IS NULL
		ORDER BY %s %s
		OFFSET $2
		LIMIT $3
		`, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}

	sems := []models.Semester{}

	for rows.Next() {
		var sem semester
		err = rows.StructScan(&sem)
		if err != nil {
			return nil, err
		}

		sems = append(sems, models.Semester{
			ID:        sem.ID,
			Name:      sem.Name,
			Type:      models.SemesterType(sem.Type),
			StartDate: sem.StartedDate,
		})
	}

	return sems, nil
}

func (r *sqlxSemesterRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*) FROM semesters 
		WHERE (name LIKE $1
		OR type::text LIKE $1 
		OR DATE(started_date)::text = $1) AND deleted_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, "%"+search+"%")

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxSemesterRepository) GetByID(ctx context.Context, ID string) (*models.Semester, error) {
	row := r.db.QueryRowxContext(ctx, "SELECT * FROM semesters WHERE id = $1 AND is_deleted = false", ID)

	var sem semester

	err := row.StructScan(&sem)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Semester not found"})
			}
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
	updateFields := getUpdateFields(req)

	query := fmt.Sprintf(`
	UPDATE semesters
	SET %s , updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	_, err := r.db.NamedExecContext(ctx, query, &semester{
		ID:          ID,
		Name:        req.Name,
		StartedDate: req.StartedDate,
		Type:        string(req.Type),
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Semester not found"})
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
