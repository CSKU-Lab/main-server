package sqlx

import (
	"context"
	"database/sql"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type courseEnrollmentRepository struct {
	db instance
}

func NewCourseEnrollmentRepository(db instance) repositories.CourseEnrollmentRepository {
	return &courseEnrollmentRepository{db: db}
}

func (r *courseEnrollmentRepository) Enroll(ctx context.Context, courseID string, studentID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO course_enrollments (course_id, student_id) VALUES ($1, $2)`,
		courseID, studentID,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return nil
		}
		return err
	}
	return nil
}

func (r *courseEnrollmentRepository) Unenroll(ctx context.Context, courseID string, studentID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM course_enrollments WHERE course_id = $1 AND student_id = $2`,
		courseID, studentID,
	)
	return err
}

func (r *courseEnrollmentRepository) IsEnrolled(ctx context.Context, courseID string, studentID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowxContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND student_id = $2)`,
		courseID, studentID,
	).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func (r *courseEnrollmentRepository) CountByCourseID(ctx context.Context, courseID string) (int, error) {
	var count int
	err := r.db.QueryRowxContext(ctx,
		`SELECT COUNT(*) FROM course_enrollments WHERE course_id = $1`,
		courseID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
