package sqlx

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/lib/pq"
)

type sqlxCourseRepository struct {
	db instance
}

func NewCourseRepository(db instance) repositories.CourseRepository {
	return &sqlxCourseRepository{db: db}
}

type course struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	Type string `db:"type"`
}

type updateCourse struct {
	ID   string  `db:"id"`
	Name *string `db:"name"`
	Type *string `db:"type"`
}

func (r *sqlxCourseRepository) Create(ctx context.Context, ID string, c *requests.CreateCourse) error {
	query := `INSERT INTO courses (id, name, type) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, ID, c.Name, c.Type)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(&cserrors.Option{
					Code:       cserrors.CourseAlreadyExists,
					HttpStatus: http.StatusInternalServerError,
					Message:    "course with that name is already exists",
				})
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to create course",
		})
	}

	return nil
}

func (r *sqlxCourseRepository) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	query := `SELECT id, name, type FROM courses WHERE id = $1 AND is_deleted = false`
	row := r.db.QueryRowxContext(ctx, query, ID)

	var course course
	err := row.StructScan(&course)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Course not found",
		})
	}

	return &models.Course{
		ID:   course.ID,
		Name: course.Name,
		Type: course.Type,
	}, nil
}

func (r *sqlxCourseRepository) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	case "all":
		archiveCondition = ""
	}

	query := fmt.Sprintf(`SELECT id, name, type FROM courses 
		WHERE LOWER(name) ILIKE $1 
		AND deleted_at IS NULL
		%s
		ORDER BY %s %s
		OFFSET $2
		LIMIT $3
		`, archiveCondition, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	courses := []models.Course{}

	for rows.Next() {
		var course course
		err = rows.StructScan(&course)
		if err != nil {
			return nil, err
		}

		courses = append(courses, models.Course{
			ID:       course.ID,
			Name:     course.Name,
			Type:     course.Type,
			Creators: make([]models.CourseCreator, 0),
		})
	}

	return courses, nil
}

func (r *sqlxCourseRepository) Count(ctx context.Context, search string, show string) (int, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	case "all":
		archiveCondition = ""
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM courses 
		WHERE LOWER(name) ILIKE $1 
		AND deleted_at IS NULL
		%s
	`, archiveCondition)

	row := r.db.QueryRowxContext(ctx, query, "%"+search+"%")

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxCourseRepository) UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error {
	fields := &updateCourse{
		ID:   ID,
		Name: c.Name,
		Type: c.Type,
	}

	updateFields := getUpdateFields(fields)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE courses
	SET %s ,updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	_, err := r.db.NamedExecContext(ctx, query, fields)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "course not found",
				})
			}
		}
		return err
	}

	return nil
}

func (r *sqlxCourseRepository) DeleteByID(ctx context.Context, ID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE courses SET is_deleted = true, deleted_at = NOW() WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
