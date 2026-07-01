package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type CreateInputSubmissionPayload struct {
	UserID             string
	NodeID             string
	DocumentMaterialID string
	LabID              string
	SectionID          *string
	Value              string
	Passed             bool
	Score              int
}

type InputSubmissionRepository interface {
	Create(ctx context.Context, payload *CreateInputSubmissionPayload) error
	// GetLatest returns the most recent submission for the given user/node/material/lab.
	// Returns (nil, nil) when there is no prior submission.
	GetLatest(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*models.InputSubmission, error)
	// ListByMaterial returns the latest submission per (user_id, node_id) for a material.
	ListByMaterial(ctx context.Context, documentMaterialID string) ([]models.InputSubmission, error)
}
