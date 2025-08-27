package sqlx

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type courseCreatorRepository struct {
	db *sqlx.DB
}

func NewCourseCreatorRepository(db *sqlx.DB) repositories.CourseCreatorRepository {
	return &courseCreatorRepository{db: db}
}

func (c *courseCreatorRepository) GetCreators(ctx context.Context, ID string) ([]models.CourseCreator, error) {
	query := `SELECT id, username, display_name, profile_image FROM course_creators CA
		  JOIN users ON users.id = CA.creator_id 
		  WHERE course_id = $1
		  ORDER BY "order" ASC`

	rows, err := c.db.QueryxContext(ctx, query, ID)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to get course creators"})
	}

	creators := []models.CourseCreator{}

	for rows.Next() {
		var creator models.CourseCreator
		err = rows.StructScan(&creator)
		if err != nil {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to scan course creator"})
		}
		creators = append(creators, creator)
	}

	return creators, nil
}

func (c *courseCreatorRepository) SetCreators(ctx context.Context, ID string, creators []string) error {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to begin transaction"})
	}
	defer tx.Rollback()

	query := `DELETE FROM course_creators WHERE course_id = $1`
	_, err = tx.ExecContext(ctx, query, ID)
	if err != nil {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to delete existing course creators"})
	}

	for order, creator := range creators {
		query := `INSERT INTO course_creators (course_id, creator_id, "order") VALUES ($1, $2, $3)`
		_, err := tx.ExecContext(ctx, query, ID, creator, order)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to set course creators"})
		}
	}

	if err = tx.Commit(); err != nil {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to commit transaction"})
	}

	return nil
}
