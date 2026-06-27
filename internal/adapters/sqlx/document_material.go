package sqlx

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type docTiptapNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs"`
	Content []docTiptapNode        `json:"content"`
}

func extractDocEmbedIDs(nodes []docTiptapNode) []string {
	var ids []string
	for _, node := range nodes {
		if node.Type == "codeMaterialEmbed" {
			if matID, ok := node.Attrs["materialID"].(string); ok && matID != "" {
				ids = append(ids, matID)
			}
		}
		ids = append(ids, extractDocEmbedIDs(node.Content)...)
	}
	return ids
}

type documentMaterialRepo struct {
	db *sqlx.DB
}

func NewDocumentMaterialRepository(db *sqlx.DB) repositories.DocumentMaterialRepository {
	return &documentMaterialRepo{db: db}
}

func (r *documentMaterialRepo) Create(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO document_materials (material_id) VALUES ($1)`, materialID)
	return err
}

func (r *documentMaterialRepo) GetByID(ctx context.Context, materialID string) (*repositories.DocumentMaterial, error) {
	var doc repositories.DocumentMaterial
	err := r.db.QueryRowxContext(ctx, `SELECT material_id, content FROM document_materials WHERE material_id = $1`, materialID).
		StructScan(&doc)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "document material not found"})
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentMaterialRepo) UpdateByID(ctx context.Context, materialID string, content string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE document_materials SET content = $2 WHERE material_id = $1`, materialID, content)
	return err
}

func (r *documentMaterialRepo) DeleteByID(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM document_materials WHERE material_id = $1`, materialID)
	return err
}

func (r *documentMaterialRepo) GetEmbeddedMaterialIDs(ctx context.Context, materialID string) ([]string, error) {
	doc, err := r.GetByID(ctx, materialID)
	if err != nil || doc.Content == nil {
		return nil, nil
	}
	var root docTiptapNode
	if err := json.Unmarshal([]byte(*doc.Content), &root); err != nil {
		return nil, nil
	}
	return extractDocEmbedIDs(root.Content), nil
}
