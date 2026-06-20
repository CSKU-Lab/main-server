package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type materialRepository struct {
	db instance
}

func NewMaterialRepository(db instance) repositories.MaterialRepository {
	return &materialRepository{db: db}
}

type materialRecord struct {
	ID                   string    `db:"id"`
	CourseID             string    `db:"course_id"`
	ForkedFromMaterialID *string   `db:"forked_from_material_id"`
	Name                 string    `db:"name"`
	Type                 string    `db:"type"`
	Visibility           string    `db:"visibility"`
	CreatedAt            time.Time `db:"created_at"`
	CreatedBy            string    `db:"created_by"`
	AutoScore            int       `db:"auto_score"`
	ManualScore          int       `db:"manual_score"`
}

func (r *materialRecord) toModel() *repositories.Material {
	return &repositories.Material{
		ID:                   r.ID,
		CourseID:             r.CourseID,
		ForkedFromMaterialID: r.ForkedFromMaterialID,
		Name:                 r.Name,
		Type:                 r.Type,
		Visibility:           r.Visibility,
		CreatedAt:            r.CreatedAt,
		CreatedBy:            r.CreatedBy,
		AutoScore:            r.AutoScore,
		ManualScore:          r.ManualScore,
	}
}

func (m *materialRepository) Create(ctx context.Context, ID string, courseID string, createdByUserID string, forkedFromMaterialID *string, req *requests.CreateMaterial) error {
	query := `INSERT INTO materials (id, course_id, forked_from_material_id, name, type, visibility, created_by, manual_score) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := m.db.ExecContext(ctx, query, ID, courseID, forkedFromMaterialID, req.Name, req.Type, req.Visibility, createdByUserID, req.ManualScore)
	if err != nil {
		return err
	}

	return nil
}

func buildVisibilityClause(visibility *repositories.VisibilityFilter, nextArgIndex int) (string, []any) {
	if visibility == nil {
		return "", nil
	}
	if visibility.OnlyPublic {
		return fmt.Sprintf(" AND visibility = $%d", nextArgIndex), []any{"public"}
	}
	return fmt.Sprintf(" AND (visibility = $%d OR created_by = $%d)", nextArgIndex, nextArgIndex+1), []any{"public", visibility.ViewerID}
}

func (m *materialRepository) GetPagination(ctx context.Context, courseID string, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter, visibility *repositories.VisibilityFilter) ([]repositories.Material, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 4)
	visibilityClause, visibilityArgs := buildVisibilityClause(visibility, len(filterArgs)+4)

	baseQuery := `SELECT id, course_id, forked_from_material_id, name, type, visibility, created_at, created_by, auto_score, manual_score FROM materials
	WHERE (name ILIKE $1 OR id IN (
		SELECT mt.material_id FROM material_tags mt
		JOIN tags t ON mt.tag_id = t.id
		WHERE LOWER(t.name) = LOWER($2)
	)) AND course_id = $3 AND is_deleted = false`

	query := fmt.Sprintf(`%s%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, visibilityClause, sortBy, sortOrder, len(filterArgs)+len(visibilityArgs)+4, len(filterArgs)+len(visibilityArgs)+5)

	args := []any{"%" + search + "%", search, courseID}
	args = append(args, filterArgs...)
	args = append(args, visibilityArgs...)
	args = append(args, (page-1)*limit, limit)

	records := []materialRecord{}
	err := m.db.SelectContext(ctx, &records, query, args...)
	if err != nil {
		return nil, err
	}

	materials := make([]repositories.Material, len(records))
	for i, record := range records {
		materials[i] = *record.toModel()
	}

	return materials, nil
}

func (m *materialRepository) Count(ctx context.Context, courseID string, search string, filters []sanitize.Filter, visibility *repositories.VisibilityFilter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 4)
	visibilityClause, visibilityArgs := buildVisibilityClause(visibility, len(filterArgs)+4)

	baseQuery := `SELECT COUNT(*) FROM materials
		WHERE (name ILIKE $1 OR id IN (
			SELECT mt.material_id FROM material_tags mt
			JOIN tags t ON mt.tag_id = t.id
			WHERE LOWER(t.name) = LOWER($2)
		)) AND course_id = $3 AND is_deleted = false`

	query := fmt.Sprintf(`%s%s%s`, baseQuery, filterWhereClause, visibilityClause)

	args := []any{"%" + search + "%", search, courseID}
	args = append(args, filterArgs...)
	args = append(args, visibilityArgs...)

	var count int
	err := m.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *materialRepository) GetByID(ctx context.Context, ID string) (*repositories.Material, error) {
	query := `SELECT id, course_id, forked_from_material_id, name, type, visibility, created_at, created_by, auto_score, manual_score FROM materials WHERE id = $1 AND is_deleted = false`

	record := &materialRecord{}
	err := m.db.GetContext(ctx, record, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Material not found"})
		}
		return nil, err
	}

	return record.toModel(), nil
}

func (m *materialRepository) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial) error {
	_, err := m.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	record := &materialUpdateRecord{
		ID:          ID,
		Name:        req.Name,
		Visibility:  req.Visibility,
		AutoScore:   req.AutoScore,
		ManualScore: req.ManualScore,
	}

	updateFields := getUpdateFields(record)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE materials
	SET %s ,updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	_, err = m.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return err
	}

	return nil
}

// materialUpdateRecord is used exclusively by the UpdateByID method.
// Score fields are *int so that getUpdateFields includes them even when the value is 0.
type materialUpdateRecord struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Visibility  string `db:"visibility"`
	AutoScore   *int   `db:"auto_score"`
	ManualScore *int   `db:"manual_score"`
}

func (m *materialRepository) DeleteByID(ctx context.Context, ID string) error {
	query := `UPDATE materials SET is_deleted = true, deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := m.db.ExecContext(ctx, query, ID)
	return err
}
