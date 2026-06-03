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
	Content     string  `json:"content"`
	MinAdjWPM   float64 `json:"min_adj_wpm"`
	MinAccuracy float64 `json:"min_accuracy"`
}

func (t *TypingMaterial) Create(ctx context.Context, matID string, _ *requests.CreateMaterial, rawReq []byte) error {
	p := &repositories.TypingMaterialPayload{}
	if rawReq != nil {
		parsed, err := parsePayload[typingMaterialPayload](rawReq)
		if err != nil {
			return err
		}
		if parsed != nil {
			p.Content = parsed.Content
			p.MinAdjWPM = parsed.MinAdjWPM
			p.MinAccuracy = parsed.MinAccuracy
		}
	}
	return t.repo.Create(ctx, matID, p)
}

func (t *TypingMaterial) GetByID(ctx context.Context, ID string) (any, error) {
	return t.repo.GetByID(ctx, ID)
}

func (t *TypingMaterial) CalculateScores(_ []byte) (*registries.MaterialScores, error) {
	return &registries.MaterialScores{AutoScore: 0, ManualScore: 0}, nil
}

func (t *TypingMaterial) UpdateByID(ctx context.Context, ID string, _ *requests.BaseUpdateMaterial, rawReq []byte) error {
	parsed, err := parsePayload[typingMaterialPayload](rawReq)
	if err != nil {
		return err
	}
	if parsed == nil {
		return nil
	}
	// Preserve existing content when update only touches scoring config
	existing, err := t.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	content := parsed.Content
	if content == "" {
		content = existing.Content
	}
	return t.repo.UpdateByID(ctx, ID, &repositories.TypingMaterialPayload{
		Content:     content,
		MinAdjWPM:   parsed.MinAdjWPM,
		MinAccuracy: parsed.MinAccuracy,
	})
}

func (t *TypingMaterial) DeleteByID(ctx context.Context, ID string) error {
	return t.repo.DeleteByID(ctx, ID)
}

func (t *TypingMaterial) Clone(ctx context.Context, sourceID string, targetID string) error {
	source, err := t.repo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}
	return t.repo.Create(ctx, targetID, &repositories.TypingMaterialPayload{
		Content:     source.Content,
		MinAdjWPM:   source.MinAdjWPM,
		MinAccuracy: source.MinAccuracy,
	})
}
