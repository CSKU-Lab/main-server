package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/lib/pq"
)

type sectionStudentRepository struct {
	db instance
}

type student struct {
	ID           string  `db:"id"`
	Username     string  `db:"username"`
	DisplayName  string  `db:"display_name"`
	ProfileImage *string `db:"profile_image"`
}

type sectionStudentSchema struct {
	SectionID string `db:"section_id"`
	StudentID string `db:"student_id"`
}

func NewSectionStudentRepository(db instance) repositories.SectionStudentRepository {
	return &sectionStudentRepository{db: db}
}

func (s *sectionStudentRepository) Add(ctx context.Context, sectionID string, studentID string) error {
	query := `INSERT INTO section_students (section_id, student_id) VALUES ($1, $2)`
	_, err := s.db.ExecContext(ctx, query, sectionID, studentID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "foreign_key_violation" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: 400,
					Message:    "Section or Student not found",
				})
			}
			if pqErr.Code.Name() == "unique_violation" {
				return cserrors.New(&cserrors.Option{
					Code:       cserrors.UniqueViolation,
					HttpStatus: 400,
					Message:    "Student already added to section",
				})
			}
		}
		return err
	}
	return nil
}

func (s *sectionStudentRepository) RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error {
	query := `DELETE FROM section_students WHERE section_id = $1 AND student_id = $2`
	_, err := s.db.ExecContext(ctx, query, sectionID, studentID)
	if err != nil {
		return err
	}
	return nil
}

func (s *sectionStudentRepository) GetBySectionID(ctx context.Context, sectionID string) ([]models.Student, error) {
	query := `SELECT id, username, display_name, profile_image FROM section_students ss
		  JOIN users u ON ss.student_id = u.id
		  WHERE ss.section_id = $1`

	dbStudents := []student{}
	err := s.db.SelectContext(ctx, &dbStudents, query, sectionID)
	if err != nil {
		return nil, err
	}

	students := make([]models.Student, 0, len(dbStudents))
	for _, dbStudent := range dbStudents {
		students = append(students, models.Student{
			ID:           dbStudent.ID,
			Username:     dbStudent.Username,
			DisplayName:  dbStudent.DisplayName,
			ProfileImage: dbStudent.ProfileImage,
		})
	}

	return students, nil
}

func (s *sectionStudentRepository) GetByStudentID(ctx context.Context, studentID string) ([]models.Section, error) {
	query := `SELECT id, name, banner, semester_id, course_id FROM section_students ss
		  JOIN sections s ON ss.section_id = s.id
		  WHERE ss.student_id = $1`

	dbSections := []sectionSchema{}
	err := s.db.SelectContext(ctx, &dbSections, query, studentID)
	if err != nil {
		return nil, err
	}

	sections := make([]models.Section, 0, len(dbSections))
	for _, dbSec := range dbSections {
		sections = append(sections, models.Section{
			ID:       dbSec.ID,
			Name:     dbSec.Name,
			CourseID: dbSec.CourseID,
		})
	}

	return sections, nil
}

func (s *sectionStudentRepository) GetBySectionAndStudentID(ctx context.Context, sectionID string, userID string) (*models.SectionStudent, error) {
	query := `SELECT student_id, section_id FROM section_students WHERE student_id = $1 AND section_id = $2`

	sectionStudentSchema := &sectionStudentSchema{}
	err := s.db.GetContext(ctx, sectionStudentSchema, query, userID, sectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "SectionStudent not found"})
		}
		return nil, err
	}

	return &models.SectionStudent{
		SectionID: sectionStudentSchema.SectionID,
		StudentID: sectionStudentSchema.StudentID,
	}, nil
}

func (s *sectionStudentRepository) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT COUNT(*) FROM section_students WHERE true`

	query := baseQuery + filterWhereClause
	fmt.Println("Count Query:", query)
	var count int
	err := s.db.GetContext(ctx, &count, query, filterArgs...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *sectionStudentRepository) GetSectionsPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]repositories.RawSection, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT id, name, banner, course_id, semester_id, created_at FROM section_students ss
		  JOIN sections s ON ss.section_id = s.id
		  WHERE s.is_deleted = false`
	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+1, len(filterArgs)+2)

	args := make([]any, 0, len(filterArgs)+2)
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	sectionsSchema := []rawSectionSchema{}
	err := s.db.SelectContext(ctx, &sectionsSchema, query, args...)
	if err != nil {
		return nil, err
	}
	sections := make([]repositories.RawSection, 0, len(sectionsSchema))
	for _, sec := range sectionsSchema {
		sections = append(sections, repositories.RawSection{
			ID:         sec.ID,
			Name:       sec.Name,
			Banner:     sec.Banner,
			CourseID:   sec.CourseID,
			SemesterID: sec.SemesterID,
			CreatedAt:  sec.CreatedAt,
		})
	}
	return sections, nil
}
