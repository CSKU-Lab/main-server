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
	ID            string  `db:"id"`
	Name          string  `db:"name"`
	Description   *string `db:"description"`
	Banner        *string `db:"banner"`
	Visibility    string  `db:"visibility"`
	TotalStudents int     `db:"total_students"`
}

type updateCourse struct {
	ID          string  `db:"id"`
	Name        *string `db:"name"`
	Description *string `db:"description"`
	Banner      *string `db:"banner"`
	Visibility  *string `db:"visibility"`
}

func (r *sqlxCourseRepository) Create(ctx context.Context, ID string, c *requests.CreateCourse) error {
	query := `INSERT INTO courses (id, name, description, banner, visibility) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, ID, c.Name, c.Description, c.Banner, c.Visibility)
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
	query := `
		SELECT id, name, description, banner, visibility,
			CASE WHEN visibility = 'public'
				THEN (SELECT COUNT(*) FROM course_enrollments WHERE course_id = courses.id)
				ELSE (SELECT COUNT(DISTINCT ss.student_id) FROM section_students ss JOIN sections s ON ss.section_id = s.id WHERE s.course_id = courses.id AND s.is_deleted = false)
			END AS total_students
		FROM courses WHERE id = $1 AND is_deleted = false`
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
		ID:            course.ID,
		Name:          course.Name,
		Description:   course.Description,
		Banner:        course.Banner,
		Visibility:    course.Visibility,
		TotalStudents: course.TotalStudents,
	}, nil
}

func (r *sqlxCourseRepository) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string, visibility string) ([]models.Course, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	case "all":
		archiveCondition = ""
	}

	visibilityCondition := ""
	args := []any{"%" + search + "%"}
	if visibility != "" {
		args = append(args, visibility)
		visibilityCondition = fmt.Sprintf("AND visibility = $%d", len(args))
	}
	args = append(args, (page-1)*pageSize, pageSize)

	query := fmt.Sprintf(`
		SELECT id, name, description, banner, visibility,
			CASE WHEN visibility = 'public'
				THEN (SELECT COUNT(*) FROM course_enrollments WHERE course_id = courses.id)
				ELSE (SELECT COUNT(DISTINCT ss.student_id) FROM section_students ss JOIN sections s ON ss.section_id = s.id WHERE s.course_id = courses.id AND s.is_deleted = false)
			END AS total_students
		FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d
		`, archiveCondition, visibilityCondition, sortBy, sortOrder, len(args)-1, len(args))

	rows, err := r.db.QueryxContext(ctx, query, args...)
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
			ID:            course.ID,
			Name:          course.Name,
			Description:   course.Description,
			Banner:        course.Banner,
			Visibility:    course.Visibility,
			TotalStudents: course.TotalStudents,
			Creators:      make([]models.CourseCreator, 0),
		})
	}

	return courses, nil
}

func (r *sqlxCourseRepository) Count(ctx context.Context, search string, show string, visibility string) (int, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	case "all":
		archiveCondition = ""
	}

	visibilityCondition := ""
	args := []any{"%" + search + "%"}
	if visibility != "" {
		args = append(args, visibility)
		visibilityCondition = fmt.Sprintf("AND visibility = $%d", len(args))
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		%s
	`, archiveCondition, visibilityCondition)

	row := r.db.QueryRowxContext(ctx, query, args...)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxCourseRepository) GetFeatured(ctx context.Context, limit int) ([]models.Course, error) {
	query := `
		SELECT id, name, description, banner, visibility,
			CASE WHEN visibility = 'public'
				THEN (SELECT COUNT(*) FROM course_enrollments WHERE course_id = courses.id)
				ELSE (SELECT COUNT(DISTINCT ss.student_id) FROM section_students ss JOIN sections s ON ss.section_id = s.id WHERE s.course_id = courses.id AND s.is_deleted = false)
			END AS total_students
		FROM courses
		WHERE visibility = 'public' AND is_archived = false AND deleted_at IS NULL
		ORDER BY RANDOM()
		LIMIT $1
	`

	rows, err := r.db.QueryxContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	courses := []models.Course{}
	for rows.Next() {
		var c course
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		courses = append(courses, models.Course{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Banner:        c.Banner,
			Visibility:    c.Visibility,
			TotalStudents: c.TotalStudents,
			Creators:      make([]models.CourseCreator, 0),
		})
	}

	return courses, nil
}

func (r *sqlxCourseRepository) GetPaginationForStudent(ctx context.Context, studentID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	}

	query := fmt.Sprintf(`
		SELECT
			CASE WHEN visibility = 'public'
				THEN courses.id
				ELSE (
					SELECT s.id FROM section_students ss
					JOIN sections s ON ss.section_id = s.id
					WHERE s.course_id = courses.id AND ss.student_id = $2 AND s.is_deleted = false
					LIMIT 1
				)
			END AS id,
			name, description, banner, visibility,
			CASE WHEN visibility = 'public'
				THEN (SELECT COUNT(*) FROM course_enrollments WHERE course_id = courses.id)
				ELSE (SELECT COUNT(DISTINCT ss.student_id) FROM section_students ss JOIN sections s ON ss.section_id = s.id WHERE s.course_id = courses.id AND s.is_deleted = false)
			END AS total_students
		FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		AND (
			id IN (SELECT course_id FROM course_enrollments WHERE student_id = $2)
			OR id IN (
				SELECT DISTINCT s.course_id FROM section_students ss
				JOIN sections s ON ss.section_id = s.id
				WHERE ss.student_id = $2 AND s.is_deleted = false
			)
		)
		ORDER BY %s %s
		OFFSET $3
		LIMIT $4
	`, archiveCondition, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", studentID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	courses := []models.Course{}
	for rows.Next() {
		var c course
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		courses = append(courses, models.Course{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Banner:        c.Banner,
			Visibility:    c.Visibility,
			TotalStudents: c.TotalStudents,
			Creators:      make([]models.CourseCreator, 0),
		})
	}

	return courses, nil
}

func (r *sqlxCourseRepository) CountForStudent(ctx context.Context, studentID string, search string, show string) (int, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		AND (
			id IN (SELECT course_id FROM course_enrollments WHERE student_id = $2)
			OR id IN (
				SELECT DISTINCT s.course_id FROM section_students ss
				JOIN sections s ON ss.section_id = s.id
				WHERE ss.student_id = $2 AND s.is_deleted = false
			)
		)
	`, archiveCondition)

	var count int
	err := r.db.QueryRowxContext(ctx, query, "%"+search+"%", studentID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxCourseRepository) GetPaginationForInstructor(ctx context.Context, instructorID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	}

	query := fmt.Sprintf(`
		SELECT
			id, name, description, banner, visibility,
			(SELECT COUNT(DISTINCT ss.student_id) FROM section_students ss
			 JOIN sections s ON ss.section_id = s.id
			 WHERE s.course_id = courses.id AND s.is_deleted = false) AS total_students
		FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		AND (
			id IN (SELECT course_id FROM course_creators WHERE creator_id = $2)
			OR id IN (
				SELECT DISTINCT s.course_id FROM section_instructors si
				JOIN sections s ON si.section_id = s.id
				WHERE si.instructor_id = $2 AND s.is_deleted = false
			)
		)
		ORDER BY %s %s
		OFFSET $3
		LIMIT $4
	`, archiveCondition, sortBy, sortOrder)

	rows, err := r.db.QueryxContext(ctx, query, "%"+search+"%", instructorID, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	courses := []models.Course{}
	for rows.Next() {
		var c course
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		courses = append(courses, models.Course{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Banner:        c.Banner,
			Visibility:    c.Visibility,
			TotalStudents: c.TotalStudents,
			Creators:      make([]models.CourseCreator, 0),
		})
	}

	return courses, nil
}

func (r *sqlxCourseRepository) CountForInstructor(ctx context.Context, instructorID string, search string, show string) (int, error) {
	var archiveCondition string
	switch show {
	case "active":
		archiveCondition = "AND is_archived = false"
	case "archived":
		archiveCondition = "AND is_archived = true"
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM courses
		WHERE LOWER(name) ILIKE $1
		AND deleted_at IS NULL
		%s
		AND (
			id IN (SELECT course_id FROM course_creators WHERE creator_id = $2)
			OR id IN (
				SELECT DISTINCT s.course_id FROM section_instructors si
				JOIN sections s ON si.section_id = s.id
				WHERE si.instructor_id = $2 AND s.is_deleted = false
			)
		)
	`, archiveCondition)

	var count int
	err := r.db.QueryRowxContext(ctx, query, "%"+search+"%", instructorID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlxCourseRepository) UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error {
	fields := &updateCourse{
		ID:          ID,
		Name:        c.Name,
		Description: c.Description,
		Banner:      c.Banner,
		Visibility:  c.Visibility,
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
