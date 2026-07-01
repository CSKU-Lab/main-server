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
			(user_id, node_id, document_material_id, lab_id, section_id, value, passed, score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		payload.UserID, payload.NodeID, payload.DocumentMaterialID, payload.LabID,
		payload.SectionID, payload.Value, payload.Passed, payload.Score,
	)
	return err
}

func (r *inputSubmissionRepo) GetLatest(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*models.InputSubmission, error) {
	var rec models.InputSubmission
	err := r.db.GetContext(ctx, &rec, `
		SELECT id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, created_at
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

func (r *inputSubmissionRepo) ListByMaterial(ctx context.Context, documentMaterialID string) ([]models.InputSubmission, error) {
	var recs []models.InputSubmission
	err := r.db.SelectContext(ctx, &recs, `
		SELECT DISTINCT ON (user_id, node_id)
			id, user_id, node_id, document_material_id, lab_id, section_id, value, passed, score, created_at
		FROM input_submissions
		WHERE document_material_id = $1
		ORDER BY user_id, node_id, created_at DESC
	`, documentMaterialID)
	if err != nil {
		return nil, err
	}
	return recs, nil
}
