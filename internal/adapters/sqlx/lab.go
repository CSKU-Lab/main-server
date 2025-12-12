package sqlx

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
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
