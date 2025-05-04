package sqlx

import (
	"context"
	"errors"
	"fmt"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type sqlxCourseRepository struct {
	db *sqlx.DB
}

func NewSqlxCourseRepository(db *sqlx.DB) repositories.CourseRepository {
	return &sqlxCourseRepository{db: db}
}

func (r *sqlxCourseRepository) Create(ctx context.Context, ID string, c *requests.Course, userID string) (*models.Course, error) {
	query := `INSERT INTO courses (id, name, code, created_by) VALUES ($1, $2, $3, $4) RETURNING *`
	row := r.db.QueryRowxContext(ctx, query, ID, c.Name, c.Code, userID)

	var course models.Course
	err := row.StructScan(&course)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Course already exists")
			}
		}
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to create course")
	}

	return &course, nil
}

func (r *sqlxCourseRepository) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	query := `SELECT * FROM courses WHERE id = $1 AND is_deleted = false`
	row := r.db.QueryRowxContext(ctx, query, ID)

	var course models.Course
	err := row.StructScan(&course)
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Course not found")
	}

	return &course, nil
}

func (r *sqlxCourseRepository) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string) ([]models.Course, error) {
	query := fmt.Sprintf(`SELECT * FROM courses 
		WHERE (
		name LIKE $1 
		OR code LIKE $1
		)
		AND deleted_at IS NULL
		ORDER BY %s %s
		OFFSET $2
		LIMIT $3
		`, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	courses := []models.Course{}

	for rows.Next() {
		var course models.Course
		err = rows.StructScan(&course)
		if err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}

	return courses, nil
}

func (r *sqlxCourseRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*) FROM courses 
		WHERE (
		name LIKE $1 
		OR code LIKE $1
		)
		AND deleted_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, "%"+search+"%")

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxCourseRepository) UpdateByID(ctx context.Context, ID string, c *requests.Course) (*models.Course, error) {
	updateFields := getUpdateFields(c)

	query := fmt.Sprintf(`
	UPDATE courses
	SET %s , updated_at = NOW()
	WHERE id = :id
	RETURNING *
	`, updateFields)

	row, err := r.db.NamedQueryContext(ctx, query, &models.Course{
		ID:   ID,
		Code: c.Code,
		Name: c.Name,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Course not found")
			}
		}
		return nil, err
	}

	var updatedCourse models.Course
	for row.Next() {
		err = row.StructScan(&updatedCourse)
		if err != nil {
			return nil, err
		}
	}

	return &updatedCourse, nil
}

func (r *sqlxCourseRepository) DeleteByID(ctx context.Context, ID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE courses SET is_deleted = true, deleted_at = NOW() WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
