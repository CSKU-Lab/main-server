package repositories

import "context"

type DocumentMaterialRepository interface {
	Create(ctx context.Context, materialID string) error
	GetByID(ctx context.Context, materialID string) (*DocumentMaterial, error)
	UpdateByID(ctx context.Context, materialID string, content string) error
	DeleteByID(ctx context.Context, materialID string) error
	// GetEmbeddedMaterialIDs parses the tiptap JSON content and returns the
	// materialIDs of all codeMaterialEmbed nodes in the document. Returns an
	// empty slice (not an error) when the document has no content or no embeds.
	GetEmbeddedMaterialIDs(ctx context.Context, materialID string) ([]string, error)
}

type DocumentMaterial struct {
	MaterialID string  `db:"material_id"`
	Content    *string `db:"content"`
}
