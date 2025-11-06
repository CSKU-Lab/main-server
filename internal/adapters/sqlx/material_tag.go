package sqlx

import (
	"context"
	"net/http"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type writeMaterialTagRepository struct {
	db instance
}

func newWriteMaterialTagRepository(db instance) repositories.WriteMaterialTagRepository {
	return &writeMaterialTagRepository{db: db}
}

func (r *writeMaterialTagRepository) SetTags(ctx context.Context, materialID string, tags []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM material_tags WHERE material_id = $1`, materialID); err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to clear existing material tags",
		})
	}

	seen := make(map[string]struct{})
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		tagKey := strings.ToLower(tag)
		if _, ok := seen[tagKey]; ok {
			continue
		}
		seen[tagKey] = struct{}{}

		tagID, err := upsertTag(ctx, r.db, tag)
		if err != nil {
			return err
		}

		if _, err := r.db.ExecContext(ctx,
			`INSERT INTO material_tags (material_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			materialID,
			tagID,
		); err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Failed to assign material tag",
			})
		}
	}

	return nil
}

func upsertTag(ctx context.Context, db instance, tag string) (string, error) {
	tagID, err := uuid.NewV7()
	if err != nil {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to generate tag UUID",
		})
	}

	var id string
	err = db.GetContext(ctx, &id, `
		INSERT INTO tags (id, name)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, tagID.String(), tag)
	if err != nil {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to upsert tag",
		})
	}

	return id, nil
}

type readMaterialTagRepository struct {
	db *sqlx.DB
}

func NewReadMaterialTagRepository(db *sqlx.DB) repositories.ReadMaterialTagRepository {
	return &readMaterialTagRepository{db: db}
}

func (r *readMaterialTagRepository) GetTags(ctx context.Context, materialID string) ([]string, error) {
	query := `SELECT t.name FROM tags t
	JOIN material_tags mt ON t.id = mt.tag_id
	WHERE mt.material_id = $1
	ORDER BY t.name ASC`

	var tags []string
	if err := r.db.SelectContext(ctx, &tags, query, materialID); err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to get material tags",
		})
	}

	return tags, nil
}
