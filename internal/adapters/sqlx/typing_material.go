package sqlx

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type typingMaterialRepo struct {
	db *sqlx.DB
}

func NewTypingMaterialRepository(db *sqlx.DB) repositories.TypingMaterialRepository {
	return &typingMaterialRepo{db: db}
}

func (r *typingMaterialRepo) Create(ctx context.Context, materialID string, payload *repositories.TypingMaterialPayload) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO typing_materials (material_id, content, min_adj_wpm, min_accuracy) VALUES ($1, $2, $3, $4)`,
		materialID, payload.Content, payload.MinAdjWPM, payload.MinAccuracy,
	)
	return err
}

func (r *typingMaterialRepo) GetByID(ctx context.Context, materialID string) (*models.TypingMaterial, error) {
	var rec struct {
		Content     string  `db:"content"`
		MinAdjWPM   float64 `db:"min_adj_wpm"`
		MinAccuracy float64 `db:"min_accuracy"`
	}
	err := r.db.QueryRowxContext(ctx,
		`SELECT content, min_adj_wpm, min_accuracy FROM typing_materials WHERE material_id = $1`, materialID,
	).StructScan(&rec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Typing material not found"})
		}
		return nil, err
	}
	return &models.TypingMaterial{
		Content:     rec.Content,
		MinAdjWPM:   rec.MinAdjWPM,
		MinAccuracy: rec.MinAccuracy,
	}, nil
}

func (r *typingMaterialRepo) UpdateByID(ctx context.Context, materialID string, payload *repositories.TypingMaterialPayload) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE typing_materials SET content = $2, min_adj_wpm = $3, min_accuracy = $4 WHERE material_id = $1`,
		materialID, payload.Content, payload.MinAdjWPM, payload.MinAccuracy,
	)
	return err
}

func (r *typingMaterialRepo) DeleteByID(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM typing_materials WHERE material_id = $1`, materialID)
	return err
}
