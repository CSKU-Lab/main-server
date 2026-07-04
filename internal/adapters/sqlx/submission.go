package sqlx

import (
	"context"
	"fmt"
	"log"
	"time"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/lib/pq"
)

type submissionRepository struct {
	db instance
}

func NewSubmissionRepository(db instance) repositories.SubmissionRepository {
	return &submissionRepository{
		db: db,
	}
}

func (s *submissionRepository) Create(ctx context.Context, payload *repositories.Submission) error {
	ipAddress := ctx.Value(contextkeys.UserKey).(contextkeys.User).IP_Address

	query := `INSERT INTO submissions
	(id, user_id, material_id, lab_id, section_id, course_id, status, submission_order, created_at, updated_at, ip_address)
	SELECT $1, $2, $3, $4, $5, $6, 'queued', COALESCE(MAX(submission_order), 0) + 1, NOW(), NOW(), $7
	FROM submissions
	WHERE user_id = $2 AND material_id = $3`

	_, err := s.db.ExecContext(ctx, query, payload.ID, payload.UserID, payload.MaterialID, payload.LabID, payload.SectionID, payload.CourseID, ipAddress)
	if err != nil {
		return err
	}
	return nil
}

func (s *submissionRepository) Update(ctx context.Context, req *repositories.UpdateSubmissionRequest) error {
	record := &submissionUpdate{
		ID:          req.ID,
		Status:      req.Status,
		AutoScore:   req.AutoScore,
		ManualScore: req.ManualScore,
	}

	updateFields := getUpdateFields(record)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE submissions
		SET %s, updated_at = NOW()
		WHERE id = :id
	`, updateFields)

	_, err := s.db.NamedExecContext(ctx, query, record)
	return err
}

// submissionUpdate is used exclusively by the Update method.
// Score fields are *int so that getUpdateFields includes them even when the value is 0.
type submissionUpdate struct {
	ID          string                   `db:"id"`
	Status      *models.SubmissionStatus `db:"status"`
	ManualScore *int                     `db:"manual_score"`
	AutoScore   *int                     `db:"auto_score"`
}

type submission struct {
	ID         string                  `db:"id"`
	Status     models.SubmissionStatus `db:"status"`
	UserID     string                  `db:"user_id"`
	LabID      string                  `db:"lab_id"`
	MaterialID string                  `db:"material_id"`
	SectionID  *string                 `db:"section_id"`
	CourseID   *string                 `db:"course_id"`
	Order      int                     `db:"submission_order"`
	CreatedAt  time.Time               `db:"created_at"`
	UpdatedAt  time.Time               `db:"updated_at"`

	IPAddress   string `db:"ip_address"`
	ManualScore int    `db:"manual_score"`
	AutoScore   int    `db:"auto_score"`
}

func (s *submissionRepository) Get(ctx context.Context, id string) (*repositories.Submission, error) {
	query := `SELECT id, user_id, lab_id, section_id, course_id, material_id, status, submission_order, auto_score, created_at
              FROM submissions
              WHERE id = $1`

	submission := submission{}
	err := s.db.GetContext(ctx, &submission, query, id)
	if err != nil {
		return nil, err
	}

	model := &repositories.Submission{
		ID:         submission.ID,
		UserID:     submission.UserID,
		LabID:      submission.LabID,
		SectionID:  submission.SectionID,
		CourseID:   submission.CourseID,
		MaterialID: submission.MaterialID,
		Status:     submission.Status,
		Order:      submission.Order,
		AutoScore:  submission.AutoScore,
		CreatedAt:  submission.CreatedAt,
	}

	return model, nil
}

func (s *submissionRepository) GetPagination(
	ctx context.Context,
	userID string,
	materialID string,
	labID string,
	sectionID string,
	page int,
	pageSize int,
	sortOrder string,
) ([]repositories.Submission, error) {
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	baseQuery := `
		SELECT id, user_id, lab_id, section_id, course_id, material_id, status, submission_order, auto_score, created_at
		FROM submissions
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if materialID != "" {
		baseQuery += fmt.Sprintf(" AND material_id = $%d", argIndex)
		args = append(args, materialID)
		argIndex++
	}

	if labID != "" {
		baseQuery += fmt.Sprintf(" AND lab_id = $%d", argIndex)
		args = append(args, labID)
		argIndex++
	}

	if sectionID != "" {
		baseQuery += fmt.Sprintf(" AND section_id = $%d", argIndex)
		args = append(args, sectionID)
		argIndex++
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at %s", sortOrder)

	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" OFFSET $%d LIMIT $%d", argIndex, argIndex+1)
	args = append(args, offset, pageSize)

	rows := []submission{}
	err := s.db.SelectContext(ctx, &rows, baseQuery, args...)
	if err != nil {
		return nil, err
	}

	result := make([]repositories.Submission, len(rows))
	for i, row := range rows {
		result[i] = repositories.Submission{
			ID:         row.ID,
			UserID:     row.UserID,
			LabID:      row.LabID,
			SectionID:  row.SectionID,
			CourseID:   row.CourseID,
			MaterialID: row.MaterialID,
			Status:     row.Status,
			Order:      row.Order,
			AutoScore:  row.AutoScore,
			CreatedAt:  row.CreatedAt,
		}
	}

	return result, nil
}

func (s *submissionRepository) Count(
	ctx context.Context,
	userID string,
	materialID string,
	labID string,
	sectionID string,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM submissions
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if materialID != "" {
		query += fmt.Sprintf(" AND material_id = $%d", argIndex)
		args = append(args, materialID)
		argIndex++
	}

	if labID != "" {
		query += fmt.Sprintf(" AND lab_id = $%d", argIndex)
		args = append(args, labID)
		argIndex++
	}

	if sectionID != "" {
		query += fmt.Sprintf(" AND section_id = $%d", argIndex)
		args = append(args, sectionID)
		argIndex++
	}

	var count int
	err := s.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *submissionRepository) GetLatestOfStudentIDInSectionID(ctx context.Context, sectionID, labID, materialID, studentID string) (*models.RawSubmission, error) {
	query := `SELECT id, user_id, lab_id, section_id, course_id, material_id, status, submission_order, created_at, updated_at, ip_address, manual_score, auto_score
              FROM submissions 
              WHERE section_id = $1 
	      AND lab_id = $2
	      AND material_id = $3
	      AND user_id = $4
	      ORDER BY submission_order DESC
	      LIMIT 1
	`

	submission := submission{}
	err := s.db.GetContext(ctx, &submission, query, sectionID, labID, materialID, studentID)
	if err != nil {
		return nil, err
	}

	result := &models.RawSubmission{
		ID:          submission.ID,
		UserID:      submission.UserID,
		LabID:       submission.LabID,
		MaterialID:  submission.MaterialID,
		Status:      submission.Status,
		Order:       submission.Order,
		CreatedAt:   submission.CreatedAt,
		UpdatedAt:   submission.UpdatedAt,
		IPAddress:   submission.IPAddress,
		ManualScore: submission.ManualScore,
		AutoScore:   submission.AutoScore,
		SectionID:   submission.SectionID,
	}

	return result, nil
}

func (s *submissionRepository) CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM (
			SELECT s.user_id
			FROM submissions s
			JOIN lab_materials lm ON s.material_id = lm.material_id
			WHERE lm.lab_id = $1 
			  AND s.section_id = $2 
			  AND s.status = 'passed'
			GROUP BY s.user_id
			HAVING COUNT(DISTINCT s.material_id) = (
				SELECT COUNT(*) 
				FROM lab_materials 
				WHERE lab_id = $1
			)
		) completed_students
	`
	var count int
	err := s.db.GetContext(ctx, &count, query, labID, sectionID)
	return count, err
}

func (s *submissionRepository) GetLatestByMaterialSectionAndLabID(ctx context.Context, materialID string, sectionID string, labID string) ([]models.RawSubmission, error) {
	query := `
		SELECT DISTINCT ON (s.user_id) 
			s.id, s.user_id, s.lab_id, s.section_id, s.course_id, s.material_id, 
			s.status, s.submission_order, s.created_at, s.updated_at, 
			s.ip_address, s.manual_score, s.auto_score
		FROM submissions s
		JOIN section_students ss ON s.user_id = ss.student_id
		WHERE s.material_id = $1 
		  AND s.section_id = $2 
		  AND s.lab_id = $3
		  AND ss.section_id = $2
		ORDER BY s.user_id, s.submission_order DESC
	`

	submissions := []submission{}
	err := s.db.SelectContext(ctx, &submissions, query, materialID, sectionID, labID)
	if err != nil {
		return nil, err
	}

	result := make([]models.RawSubmission, len(submissions))
	for i, row := range submissions {
		result[i] = models.RawSubmission{
			ID:          row.ID,
			UserID:      row.UserID,
			LabID:       row.LabID,
			MaterialID:  row.MaterialID,
			Status:      row.Status,
			Order:       row.Order,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			IPAddress:   row.IPAddress,
			ManualScore: row.ManualScore,
			AutoScore:   row.AutoScore,
		}
	}

	return result, nil
}

func (s *submissionRepository) GetPaginationByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.RawSubmission, error) {
	query := `
		SELECT id, user_id, lab_id, section_id, course_id, material_id, 
			   status, submission_order, created_at, updated_at, 
			   ip_address, manual_score, auto_score
		FROM submissions
		WHERE material_id = $1 
		  AND section_id = $2 
		  AND lab_id = $3
		  AND user_id = $4
		ORDER BY %s %s
		OFFSET $5 LIMIT $6
	`

	query = fmt.Sprintf(query, sortBy, sortOrder)
	offset := (page - 1) * pageSize
	log.Println(query)

	submissions := []submission{}
	err := s.db.SelectContext(ctx, &submissions, query, materialID, sectionID, labID, studentID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]models.RawSubmission, len(submissions))
	for i, row := range submissions {
		result[i] = models.RawSubmission{
			ID:          row.ID,
			UserID:      row.UserID,
			LabID:       row.LabID,
			MaterialID:  row.MaterialID,
			Status:      row.Status,
			Order:       row.Order,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			IPAddress:   row.IPAddress,
			ManualScore: row.ManualScore,
			AutoScore:   row.AutoScore,
		}
	}

	return result, nil
}

func (s *submissionRepository) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM submissions WHERE id = $1`, id)
	return err
}

func (s *submissionRepository) CountByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM submissions
		WHERE material_id = $1
		  AND section_id = $2
		  AND lab_id = $3
		  AND user_id = $4
	`

	var count int
	err := s.db.GetContext(ctx, &count, query, materialID, sectionID, labID, studentID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *submissionRepository) GetLatestScoresByMaterialsForSection(ctx context.Context, materialIDs []string, sectionID, labID string) ([]models.RawSubmission, error) {
	if len(materialIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT DISTINCT ON (user_id, material_id)
			id, user_id, material_id, lab_id, section_id, course_id,
			status, submission_order, created_at, updated_at,
			ip_address, manual_score, auto_score
		FROM submissions
		WHERE material_id = ANY($1)
		  AND section_id = $2
		  AND lab_id = $3
		ORDER BY user_id, material_id, submission_order DESC
	`

	rows := []submission{}
	err := s.db.SelectContext(ctx, &rows, query, pq.Array(materialIDs), sectionID, labID)
	if err != nil {
		return nil, err
	}

	result := make([]models.RawSubmission, len(rows))
	for i, row := range rows {
		result[i] = models.RawSubmission{
			ID:          row.ID,
			UserID:      row.UserID,
			MaterialID:  row.MaterialID,
			LabID:       row.LabID,
			Status:      row.Status,
			Order:       row.Order,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			IPAddress:   row.IPAddress,
			ManualScore: row.ManualScore,
			AutoScore:   row.AutoScore,
		}
	}
	return result, nil
}

func (s *submissionRepository) GetLatestScoresBySection(ctx context.Context, sectionID string) ([]models.RawSubmission, error) {
	query := `
		SELECT DISTINCT ON (user_id, lab_id, material_id)
			id, user_id, material_id, lab_id, section_id, course_id,
			status, submission_order, created_at, updated_at,
			ip_address, manual_score, auto_score
		FROM submissions
		WHERE section_id = $1
		ORDER BY user_id, lab_id, material_id, submission_order DESC
	`

	rows := []submission{}
	err := s.db.SelectContext(ctx, &rows, query, sectionID)
	if err != nil {
		return nil, err
	}

	result := make([]models.RawSubmission, len(rows))
	for i, row := range rows {
		result[i] = models.RawSubmission{
			ID:          row.ID,
			UserID:      row.UserID,
			MaterialID:  row.MaterialID,
			LabID:       row.LabID,
			Status:      row.Status,
			Order:       row.Order,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			IPAddress:   row.IPAddress,
			ManualScore: row.ManualScore,
			AutoScore:   row.AutoScore,
		}
	}
	return result, nil
}
