package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type TypingMaterial struct {
	repo repositories.TypingMaterialRepository
}

func NewTypingMaterial(repo repositories.TypingMaterialRepository) *TypingMaterial {
	return &TypingMaterial{repo: repo}
}

type typingMaterialPayload struct {
	Content string `json:"content"`
}

func (t *TypingMaterial) Create(ctx context.Context, matID string, _ *requests.CreateMaterial, rawReq []byte) error {
	content := ""
	if rawReq != nil {
		payload, err := parsePayload[typingMaterialPayload](rawReq)
		if err != nil {
			return err
		}
		if payload != nil {
			content = payload.Content
		}
	}
	return t.repo.Create(ctx, matID, content)
}

func (t *TypingMaterial) GetByID(ctx context.Context, ID string) (any, error) {
	return t.repo.GetByID(ctx, ID)
}

func (t *TypingMaterial) CalculateScores(_ []byte) (*registries.MaterialScores, error) {
	return &registries.MaterialScores{AutoScore: 0, ManualScore: 0}, nil
}

func (t *TypingMaterial) UpdateByID(ctx context.Context, ID string, _ *requests.BaseUpdateMaterial, rawReq []byte) error {
	payload, err := parsePayload[typingMaterialPayload](rawReq)
	if err != nil {
		return err
	}
	if payload == nil || payload.Content == "" {
		return nil
	}
	return t.repo.UpdateByID(ctx, ID, payload.Content)
}

func (t *TypingMaterial) DeleteByID(ctx context.Context, ID string) error {
	return t.repo.DeleteByID(ctx, ID)
}
