package registrables

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type DocumentMaterial struct {
	repo         repositories.DocumentMaterialRepository
	materialRepo repositories.MaterialRepository
}

func NewDocumentMaterial(repo repositories.DocumentMaterialRepository, materialRepo repositories.MaterialRepository) *DocumentMaterial {
	return &DocumentMaterial{repo: repo, materialRepo: materialRepo}
}

type DocumentMaterialPayload struct {
	Content *string `json:"content,omitempty"`
}

type DocumentMaterialResponse struct {
	Content *string `json:"content"`
}

type tiptapNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs"`
	Content []tiptapNode           `json:"content"`
}

func sumEmbeddedCodeScores(nodes []tiptapNode) int {
	total := 0
	for _, node := range nodes {
		if node.Type == "codeMaterialEmbed" {
			if score, ok := node.Attrs["autoScore"].(float64); ok {
				total += int(score)
			}
		}
		total += sumEmbeddedCodeScores(node.Content)
	}
	return total
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

func (d *DocumentMaterial) CalculateScores(rawReq []byte) (*registries.MaterialScores, error) {
	if rawReq == nil {
		return &registries.MaterialScores{}, nil
	}

	payload, err := parsePayload[DocumentMaterialPayload](rawReq)
	if err != nil {
		return nil, err
	}

	if payload == nil || payload.Content == nil {
		return &registries.MaterialScores{}, nil
	}

	var doc tiptapNode
	if err := json.Unmarshal([]byte(*payload.Content), &doc); err != nil {
		// invalid JSON content — treat as no embeds
		return &registries.MaterialScores{}, nil
	}

	autoScore := sumEmbeddedCodeScores(doc.Content)
	return &registries.MaterialScores{AutoScore: autoScore}, nil
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
