package registrables

import (
	"context"
	"math"
	"net/http"
	"time"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	typingsession "github.com/CSKU-Lab/main-server/internal/typing_session"
)

const (
	maxRawWPM   = 300.0
	minDuration = 3.0
)

type typingSubmission struct {
	repo          repositories.TypingSubmissionRepository
	typingMatRepo repositories.TypingMaterialRepository
	secret        string
}

func NewTypingSubmission(repo repositories.TypingSubmissionRepository, typingMatRepo repositories.TypingMaterialRepository, secret string) registries.SubmissionRegistrable {
	return &typingSubmission{
		repo:          repo,
		typingMatRepo: typingMatRepo,
		secret:        secret,
	}
}

type createTypingSubmissionPayload struct {
	Token     string `json:"token"`
	TypedText string `json:"typed_text"`
}

func (t *typingSubmission) Create(ctx context.Context, uow repositories.UoWInstance, submissionID string, matID string, payload []byte) error {
	parsed, err := parsePayload[createTypingSubmissionPayload](payload)
	if err != nil {
		return err
	}

	user, ok := ctx.Value(contextkeys.UserKey).(contextkeys.User)
	if !ok {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Cannot extract user from context"})
	}

	claims, err := typingsession.VerifyToken(t.secret, parsed.Token)
	if err != nil {
		return err
	}

	if claims.StudentID != user.ID {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Session token does not belong to this user"})
	}
	if claims.MaterialID != matID {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Session token does not match material"})
	}

	if len(parsed.TypedText) == 0 {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "No typed text provided"})
	}

	duration := time.Since(claims.StartedAt).Seconds()
	if duration < minDuration {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Submission rejected: duration too short"})
	}

	mat, err := t.typingMatRepo.GetByID(ctx, matID)
	if err != nil {
		return err
	}

	correct := countCorrectChars(mat.Content, parsed.TypedText)
	errors := len(parsed.TypedText) - correct
	durationMin := duration / 60.0
	rawWPM := (float64(len(parsed.TypedText)) / 5.0) / durationMin
	adjWPM := (float64(correct) / 5.0) / durationMin
	errorRate := (float64(errors) / math.Max(float64(len(parsed.TypedText)), 1)) * 100.0

	if rawWPM > maxRawWPM {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Submission rejected: typing speed exceeds human limits"})
	}

	if err := uow.TypingSubmission().Create(ctx, &repositories.CreateTypingSubmissionPayload{
		SubmissionID: submissionID,
		RawWPM:       rawWPM,
		AdjustedWPM:  adjWPM,
		ErrorRate:    errorRate,
		Duration:     duration,
	}); err != nil {
		return err
	}

	status := models.PASSED
	return uow.Submission().Update(ctx, &repositories.UpdateSubmissionRequest{
		ID:     submissionID,
		Status: &status,
	})
}

func (t *typingSubmission) Update(_ context.Context, _ repositories.UoWInstance, _ string, _ []byte) error {
	return nil
}

func (t *typingSubmission) Get(ctx context.Context, submissionID string, _ string) (any, error) {
	return t.repo.Get(ctx, submissionID)
}

func (t *typingSubmission) GetByIDs(ctx context.Context, submissionIDs []string, _ string) (map[string]any, error) {
	submissions, err := t.repo.GetByIDs(ctx, submissionIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any, len(submissions))
	for id, s := range submissions {
		result[id] = s
	}
	return result, nil
}

func (t *typingSubmission) GetOverviewsPayload(ctx context.Context, submissionIDs []string) (map[string]any, error) {
	return t.GetByIDs(ctx, submissionIDs, "")
}

func (t *typingSubmission) GetOverviewStats(payload any) any {
	return payload
}

func (t *typingSubmission) GetOverviewStatsByID(ctx context.Context, submissionID string) any {
	sub, err := t.repo.Get(ctx, submissionID)
	if err != nil {
		return nil
	}
	return sub
}

func countCorrectChars(content, typed string) int {
	correct := 0
	for i := 0; i < len(typed) && i < len(content); i++ {
		if typed[i] == content[i] {
			correct++
		}
	}
	return correct
}
