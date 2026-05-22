package registrables

import (
	"context"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type DocumentMaterial struct {
	repo repositories.DocumentMaterialRepository
}

func NewDocumentMaterial(repo repositories.DocumentMaterialRepository) *DocumentMaterial {
	return &DocumentMaterial{repo: repo}
}

type DocumentMaterialPayload struct {
	Content *string `json:"content,omitempty"`
}

type DocumentMaterialResponse struct {
	Content *string `json:"content"`
}

func (d *DocumentMaterial) Create(ctx context.Context, matID string, _ *requests.CreateMaterial, _ []byte) error {
	return d.repo.Create(ctx, matID)
}

func (d *DocumentMaterial) GetByID(ctx context.Context, ID string) (any, error) {
	doc, err := d.repo.GetByID(ctx, ID)
	if err != nil {
		var csErr *cserrors.Error
		if errors.As(err, &csErr) && csErr.HttpStatus == http.StatusNotFound {
			// row missing (material predates document_materials table) — return empty
			return &DocumentMaterialResponse{Content: nil}, nil
		}
		return nil, err
	}
	return &DocumentMaterialResponse{Content: doc.Content}, nil
}

func (d *DocumentMaterial) CalculateScores(_ []byte) (*registries.MaterialScores, error) {
	return &registries.MaterialScores{AutoScore: 0, ManualScore: 0}, nil
}

func (d *DocumentMaterial) UpdateByID(ctx context.Context, ID string, _ *requests.BaseUpdateMaterial, rawReq []byte) error {
	payload, err := parsePayload[DocumentMaterialPayload](rawReq)
	if err != nil {
		return err
	}
	if payload == nil || payload.Content == nil {
		return nil
	}
	return d.repo.UpdateByID(ctx, ID, *payload.Content)
}

func (d *DocumentMaterial) DeleteByID(ctx context.Context, ID string) error {
	return d.repo.DeleteByID(ctx, ID)
}

func (d *DocumentMaterial) Clone(ctx context.Context, sourceID string, targetID string) error {
	source, err := d.repo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}
	if err := d.repo.Create(ctx, targetID); err != nil {
		return err
	}
	if source.Content == nil {
		return nil
	}
	return d.repo.UpdateByID(ctx, targetID, *source.Content)
}
