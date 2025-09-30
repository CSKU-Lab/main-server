package repositories

import "context"

type SectionStudentRepository interface {
	Add(ctx context.Context, sectionID string, studentID string) error
	DeleteBySectionID(ctx context.Context, sectionID string) error
	// GetBySectionID(ctx context.Context, sectionID string) error
}
