package sqlx

import (
	"context"
	"database/sql"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type inputSubmissionRepo struct {
	db *sqlx.DB
}

func NewInputSubmissionRepository(db *sqlx.DB) repositories.InputSubmissionRepository {
	return &inputSubmissionRepo{db: db}
}

func (r *inputSubmissionRepo) Create(ctx context.Context, payload *repositories.CreateInputSubmissionPayload) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO input_submissions
			(user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		payload.UserID, payload.NodeID, payload.DocumentMaterialID, payload.LabID,
		payload.SectionID, payload.Value, payload.Passed, payload.Score, payload.Graded,
	)
	return err
}

func (r *inputSubmissionRepo) GetLatest(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*models.InputSubmission, error) {
	var rec models.InputSubmission
	err := r.db.GetContext(ctx, &rec, `
		SELECT id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded, created_at
		FROM input_submissions
		WHERE user_id = $1 AND node_id = $2 AND document_material_id = $3 AND lab_id = $4
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, nodeID, documentMaterialID, labID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *inputSubmissionRepo) GetByID(ctx context.Context, id string) (*models.InputSubmission, error) {
	var rec models.InputSubmission
	err := r.db.GetContext(ctx, &rec, `
		SELECT id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded, created_at
		FROM input_submissions
		WHERE id = $1
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *inputSubmissionRepo) ListByMaterial(ctx context.Context, documentMaterialID string) ([]models.InputSubmission, error) {
	var recs []models.InputSubmission
	err := r.db.SelectContext(ctx, &recs, `
		SELECT DISTINCT ON (user_id, node_id)
			id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded, created_at
		FROM input_submissions
		WHERE document_material_id = $1
		ORDER BY user_id, node_id, created_at DESC
	`, documentMaterialID)
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (r *inputSubmissionRepo) ListLatestByMaterialSectionLab(ctx context.Context, documentMaterialID, sectionID, labID string) ([]models.InputSubmission, error) {
	var recs []models.InputSubmission
	err := r.db.SelectContext(ctx, &recs, `
		SELECT DISTINCT ON (user_id, node_id)
			id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded, created_at
		FROM input_submissions
		WHERE document_material_id = $1 AND section_id = $2 AND lab_id = $3
		ORDER BY user_id, node_id, created_at DESC
	`, documentMaterialID, sectionID, labID)
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (r *inputSubmissionRepo) ListLatestBySection(ctx context.Context, sectionID string) ([]models.InputSubmission, error) {
	var recs []models.InputSubmission
	err := r.db.SelectContext(ctx, &recs, `
		SELECT DISTINCT ON (user_id, document_material_id, node_id)
			id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, graded, created_at
		FROM input_submissions
		WHERE section_id = $1
		ORDER BY user_id, document_material_id, node_id, created_at DESC
	`, sectionID)
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (r *inputSubmissionRepo) Grade(ctx context.Context, id string, score int, passed bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE input_submissions
		SET score = $2, passed = $3, graded = true
		WHERE id = $1
	`, id, score, passed)
	return err
}
