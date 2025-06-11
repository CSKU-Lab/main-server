package sqlx

import (
	"context"
	"errors"
	"fmt"
	"log"

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

func (r *sqlxCourseRepository) Create(ctx context.Context, ID string, c *requests.Course) error {
	query := `INSERT INTO courses (id, name) VALUES ($1, $2)`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to begin transaction")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
		}
	}()

	_, err = tx.ExecContext(ctx, query, ID, c.Name)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Course already exists")
			}
		}

		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to create course")
	}

	for order, creator := range c.Creators {
		query := `INSERT INTO course_creators (course_id, creator_id, "order") VALUES ($1, $2, $3)`
		_, err := tx.ExecContext(ctx, query, ID, creator, order)
		if err != nil {
			log.Printf("Failed to insert course creator: %v", err)
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to set course creators")
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to commit transaction")
	}

	return nil
}

func (r *sqlxCourseRepository) SetCreators(ctx context.Context, ID string, creators []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to begin transaction")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
		}
	}()

	query := `DELETE FROM course_creators WHERE course_id = $1`
	_, err = tx.ExecContext(ctx, query, ID)
	if err != nil {
		log.Printf("Failed to delete existing course creators: %v", err)
		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to delete existing course creators")
	}

	for order, creator := range creators {
		query := `INSERT INTO course_creators (course_id, creator_id, "order") VALUES ($1, $2, $3)`
		_, err := tx.ExecContext(ctx, query, ID, creator, order)
		if err != nil {
			log.Printf("Failed to insert course creator: %v", err)
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to set course creators")
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to commit transaction")
	}

	return nil
}

func (r *sqlxCourseRepository) GetCreators(ctx context.Context, ID string) ([]models.CourseCreator, error) {
	query := `SELECT id, display_name, profile_image FROM course_creators CA
		  JOIN users ON users.id = CA.creator_id 
		  WHERE course_id = $1
		  ORDER BY "order" ASC`

	rows, err := r.db.QueryxContext(ctx, query, ID)
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to get course creators")
	}

	creators := []models.CourseCreator{}

	for rows.Next() {
		var creator models.CourseCreator
		err = rows.StructScan(&creator)
		if err != nil {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to scan course creator")
		}
		creators = append(creators, creator)
	}

	return creators, nil
}

func (r *sqlxCourseRepository) GetByID(ctx context.Context, ID string) (*repositories.Course, error) {
	query := `SELECT * FROM courses WHERE id = $1 AND is_deleted = false`
	row := r.db.QueryRowxContext(ctx, query, ID)

	var course repositories.Course
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

		creators, err := r.GetCreators(ctx, course.ID)
		if err != nil {
			return nil, err
		}

		course.Creators = creators

		courses = append(courses, course)
	}

	return courses, nil
}

func (r *sqlxCourseRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*) FROM courses 
		WHERE (
		name LIKE $1 
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

func (r *sqlxCourseRepository) UpdateByID(ctx context.Context, ID string, c *requests.Course) error {
	updateFields := getUpdateFields(c)

	log.Println(updateFields)
	query := fmt.Sprintf(`
	UPDATE courses
	SET %s , updated_at = NOW()
	WHERE id = :id
	RETURNING *
	`, updateFields)

	row, err := r.db.NamedQueryContext(ctx, query, &models.Course{
		ID:   ID,
		Name: c.Name,
	})
	if err != nil {
		log.Println(err)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "22P02" {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Course not found")
			}
		}
		return err
	}

	var updatedCourse models.Course
	for row.Next() {
		err = row.StructScan(&updatedCourse)
		if err != nil {
			return err
		}
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
