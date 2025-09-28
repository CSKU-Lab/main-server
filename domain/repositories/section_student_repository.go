package repositories

import "context"

type SectionStudentRepository interface {
	Add(ctx context.Context, sectionID string, id string) error
	DeleteBySectionID(ctx context.Context, sectionID string) error
	// GetBySectionID(ctx context.Context, sectionID string) error
}
