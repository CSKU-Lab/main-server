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
	Graded             bool
}

type InputSubmissionRepository interface {
	Create(ctx context.Context, payload *CreateInputSubmissionPayload) error
	// GetLatest returns the most recent submission for the given user/node/material/lab.
	// Returns (nil, nil) when there is no prior submission.
	GetLatest(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*models.InputSubmission, error)
	// GetByID returns a single submission, or (nil, nil) when it does not exist.
	GetByID(ctx context.Context, id string) (*models.InputSubmission, error)
	// ListByMaterial returns the latest submission per (user_id, node_id) for a material.
	ListByMaterial(ctx context.Context, documentMaterialID string) ([]models.InputSubmission, error)
	// ListLatestByMaterialSectionLab returns the latest submission per (user_id, node_id)
	// for a document material scoped to one section and lab.
	ListLatestByMaterialSectionLab(ctx context.Context, documentMaterialID, sectionID, labID string) ([]models.InputSubmission, error)
	// ListLatestBySection returns the latest submission per (user_id, document_material_id, node_id)
	// across every document material in a section — used to build the gradebook.
	ListLatestBySection(ctx context.Context, sectionID string) ([]models.InputSubmission, error)
	// Grade sets the score/passed of a submission and marks it graded (manual mode).
	Grade(ctx context.Context, id string, score int, passed bool) error
}
