package registries

import (
	"context"
	"fmt"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type SubmissionRegistrable interface {
	Create(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, payload []byte) error
	Update(ctx context.Context, uowRepo repositories.UoWInstance, submissioID string, payload []byte) error
	Get(ctx context.Context, submissionId string) (any, error)
}

type submissionRegistry struct {
	handlers map[string]SubmissionRegistrable
}

type SubmissionRegistry interface {
	Register(key string, handler SubmissionRegistrable)
	GetHandler(key string) (SubmissionRegistrable, error)
}

func NewSubmission() SubmissionRegistry {
	return &submissionRegistry{
		handlers: make(map[string]SubmissionRegistrable),
	}
}

func (s *submissionRegistry) Register(key string, handler SubmissionRegistrable) {
	s.handlers[key] = handler
}

func (s *submissionRegistry) GetHandler(key string) (SubmissionRegistrable, error) {
	handler, exists := s.handlers[key]
	if !exists {
		return nil, fmt.Errorf("handler not found for key: %s", key)
	}
	return handler, nil
}
