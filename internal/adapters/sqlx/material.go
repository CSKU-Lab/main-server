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
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Type        string    `db:"type"`
	Visibility  string    `db:"visibility"`
	CreatedAt   time.Time `db:"created_at"`
	CreatedBy   string    `db:"created_by"`
	AutoScore   int       `db:"auto_score"`
	ManualScore int       `db:"manual_score"`
}

func (r *materialRecord) toModel() *repositories.Material {
	return &repositories.Material{
		ID:          r.ID,
		Name:        r.Name,
		Type:        r.Type,
		Visibility:  r.Visibility,
		CreatedAt:   r.CreatedAt,
		CreatedBy:   r.CreatedBy,
		AutoScore:   r.AutoScore,
		ManualScore: r.ManualScore,
	}
}

func (m *materialRepository) Create(ctx context.Context, ID string, createdByUserID string, req *requests.CreateMaterial) error {
	query := `INSERT INTO materials (id, name, type, visibility, created_by) VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.ExecContext(ctx, query, ID, req.Name, req.Type, req.Visibility, createdByUserID)
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

func (m *materialRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter, visibility *repositories.VisibilityFilter) ([]repositories.Material, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)
	visibilityClause, visibilityArgs := buildVisibilityClause(visibility, len(filterArgs)+2)

	baseQuery := `SELECT id, name, type, visibility, created_at, created_by, auto_score, manual_score FROM materials
	WHERE (name ILIKE $1)`

	query := fmt.Sprintf(`%s%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, visibilityClause, sortBy, sortOrder, len(filterArgs)+len(visibilityArgs)+2, len(filterArgs)+len(visibilityArgs)+3)

	args := []any{"%" + search + "%"}
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

func (m *materialRepository) Count(ctx context.Context, search string, filters []sanitize.Filter, visibility *repositories.VisibilityFilter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)
	visibilityClause, visibilityArgs := buildVisibilityClause(visibility, len(filterArgs)+2)

	baseQuery := `SELECT COUNT(*) FROM materials
		WHERE (name ILIKE $1)`

	query := fmt.Sprintf(`%s%s%s`, baseQuery, filterWhereClause, visibilityClause)

	args := []any{"%" + search + "%"}
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
	query := `SELECT id, name, type, visibility, created_at, created_by, auto_score, manual_score FROM materials WHERE id = $1`

	record := &materialRecord{}
	err := m.db.GetContext(ctx, record, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Material not found"})
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
	query := `DELETE FROM materials WHERE id = $1`
	_, err := m.db.ExecContext(ctx, query, ID)
	return err
}
