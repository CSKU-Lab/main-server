package registrables

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	typingsession "github.com/CSKU-Lab/main-server/internal/typing_session"
)

const (
	maxRawWPM = 300.0
)

type TypingSubmission struct {
	repo          repositories.TypingSubmissionRepository
	typingMatRepo repositories.TypingMaterialRepository
	materialRepo  repositories.MaterialRepository
	secret        string
}

func NewTypingSubmission(repo repositories.TypingSubmissionRepository, typingMatRepo repositories.TypingMaterialRepository, materialRepo repositories.MaterialRepository, secret string) *TypingSubmission {
	return &TypingSubmission{
		repo:          repo,
		typingMatRepo: typingMatRepo,
		materialRepo:  materialRepo,
		secret:        secret,
	}
}

type keystroke struct {
	K string `json:"k"` // key pressed ("Backspace" or single char)
	T int64  `json:"t"` // ms since first keystroke
}

type createTypingSubmissionPayload struct {
	Token      string      `json:"token"`
	Keystrokes []keystroke `json:"keystrokes"`
}

// replayKeystrokes replays keystrokes against content and returns:
//   - finalText: resulting text after applying all keystrokes
//   - totalErrors: number of incorrect char keystrokes at the time of pressing
//   - totalCharKeystrokes: total non-Backspace keystrokes (for rawWPM)
func replayKeystrokes(content string, keystrokes []keystroke) (finalText string, totalErrors, totalCharKeystrokes int, err error) {
	contentRunes := []rune(content)
	buf := make([]rune, 0, len(contentRunes))

	for i, ks := range keystrokes {
		if i > 0 && ks.T < keystrokes[i-1].T {
			return "", 0, 0, errors.New("keystroke timestamps are not monotonically increasing")
		}

		if ks.K == "Backspace" {
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
			continue
		}

		r := []rune(ks.K)
		if len(r) != 1 {
			return "", 0, 0, fmt.Errorf("invalid keystroke: %q", ks.K)
		}

		cursor := len(buf)
		if cursor < len(contentRunes) && r[0] != contentRunes[cursor] {
			totalErrors++
		}
		buf = append(buf, r[0])
		totalCharKeystrokes++
	}

	finalText = string(buf)
	return
}

func (t *TypingSubmission) Create(ctx context.Context, uow repositories.UoWInstance, submissionID string, matID string, payload []byte) error {
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
	if claims.TypingStartedAt.IsZero() {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Typing session was not started"})
	}

	if len(parsed.Keystrokes) == 0 {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "No keystrokes provided"})
	}

	mat, err := t.typingMatRepo.GetByID(ctx, matID)
	if err != nil {
		return err
	}

	finalText, totalErrors, totalCharKeystrokes, err := replayKeystrokes(mat.Content, parsed.Keystrokes)
	if err != nil {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: err.Error()})
	}

	// Validate the replayed text matches the material content exactly (anti-spoof)
	if finalText != mat.Content {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Typed text does not match material content"})
	}

	receivedAt := time.Now()
	duration := receivedAt.Sub(claims.TypingStartedAt).Seconds()

	// Minimum duration: time to type the full text at the maximum allowed WPM
	minDuration := (float64(len([]rune(mat.Content))) / 5.0) / maxRawWPM * 60.0
	if duration < minDuration {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Submission rejected: duration too short"})
	}

	contentLen := float64(len([]rune(mat.Content)))
	durationMin := duration / 60.0
	rawWPM := (float64(totalCharKeystrokes) / 5.0) / durationMin
	adjWPM := (contentLen / 5.0) / durationMin // finalText == content so all chars are correct
	errorRate := (float64(totalErrors) / math.Max(contentLen, 1)) * 100.0

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

	baseMat, err := t.materialRepo.GetByID(ctx, matID)
	if err != nil {
		return err
	}

	accuracy := 100.0 - errorRate
	autoScore := 0
	status := models.PASSED

	if baseMat.AutoScore > 0 {
		wpmOK := mat.MinAdjWPM == 0 || adjWPM >= mat.MinAdjWPM
		accOK := mat.MinAccuracy == 0 || accuracy >= mat.MinAccuracy
		if wpmOK && accOK {
			autoScore = baseMat.AutoScore
		} else {
			status = models.FAILED
		}
	}

	return uow.Submission().Update(ctx, &repositories.UpdateSubmissionRequest{
		ID:        submissionID,
		Status:    &status,
		AutoScore: &autoScore,
	})
}

func (t *TypingSubmission) Update(_ context.Context, _ repositories.UoWInstance, _ string, _ []byte) error {
	return nil
}

func (t *TypingSubmission) Get(ctx context.Context, submissionID string, _ string) (any, error) {
	return t.repo.Get(ctx, submissionID)
}

func (t *TypingSubmission) GetByIDs(ctx context.Context, submissionIDs []string, _ string) (map[string]any, error) {
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

func (t *TypingSubmission) GetOverviewsPayload(ctx context.Context, submissionIDs []string) (map[string]any, error) {
	return t.GetByIDs(ctx, submissionIDs, "")
}

func (t *TypingSubmission) GetOverviewStats(payload any) any {
	return payload
}

func (t *TypingSubmission) GetOverviewStatsByID(ctx context.Context, submissionID string) any {
	sub, err := t.repo.Get(ctx, submissionID)
	if err != nil {
		return nil
	}
	return sub
}
