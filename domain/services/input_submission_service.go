package services

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"go.uber.org/zap"
)

type SubmitInputAnswerInput struct {
	UserID             string
	NodeID             string
	DocumentMaterialID string
	LabID              string
	SectionID          *string
	Value              string
}

type SubmitInputAnswerResult struct {
	Passed bool `json:"passed"`
	Score  int  `json:"score"`
	// Graded is false for manual-mode submissions awaiting instructor grading.
	Graded bool `json:"graded"`
}

type GradeInputInput struct {
	SubmissionID string
	Score        int
}

type InputResult struct {
	Submitted bool   `json:"submitted"`
	Passed    bool   `json:"passed"`
	Score     int    `json:"score"`
	Graded    bool   `json:"graded"`
	Value     string `json:"value"`
}

type InputSubmissionResult struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	NodeID    string    `json:"node_id"`
	Value     string    `json:"value"`
	Passed    bool      `json:"passed"`
	Score     int       `json:"score"`
	Graded    bool      `json:"graded"`
	CreatedAt time.Time `json:"created_at"`
}

type InputSubmissionService interface {
	SubmitInputAnswer(ctx context.Context, input *SubmitInputAnswerInput) (*SubmitInputAnswerResult, error)
	GetMyLatestResult(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*InputResult, error)
	ListByMaterial(ctx context.Context, documentMaterialID string) ([]InputSubmissionResult, error)
	// GradeManualInput sets an instructor-assigned score on a manual-mode submission.
	GradeManualInput(ctx context.Context, input *GradeInputInput) error
}

type inputSubmissionService struct {
	repo    repositories.InputSubmissionRepository
	docRepo repositories.DocumentMaterialRepository
	logger  *zap.SugaredLogger
}

func NewInputSubmissionService(
	repo repositories.InputSubmissionRepository,
	docRepo repositories.DocumentMaterialRepository,
	logger *zap.SugaredLogger,
) InputSubmissionService {
	return &inputSubmissionService{
		repo:    repo,
		docRepo: docRepo,
		logger:  logger,
	}
}

type inputTiptapNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs"`
	Content []inputTiptapNode      `json:"content"`
}

// findInputEmbed walks the tiptap tree and returns the inputEmbed node whose
// attrs.nodeID matches the given nodeID. Returns nil when not found.
func findInputEmbed(nodes []inputTiptapNode, nodeID string) *inputTiptapNode {
	for i := range nodes {
		node := &nodes[i]
		if node.Type == "inputEmbed" {
			if id, ok := node.Attrs["nodeID"].(string); ok && id == nodeID {
				return node
			}
		}
		if found := findInputEmbed(node.Content, nodeID); found != nil {
			return found
		}
	}
	return nil
}

func (s *inputSubmissionService) SubmitInputAnswer(ctx context.Context, input *SubmitInputAnswerInput) (*SubmitInputAnswerResult, error) {
	doc, err := s.docRepo.GetByID(ctx, input.DocumentMaterialID)
	if err != nil {
		return nil, err
	}
	if doc.Content == nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "input field not found"})
	}

	var root inputTiptapNode
	if err := json.Unmarshal([]byte(*doc.Content), &root); err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "invalid document content"})
	}

	node := findInputEmbed(root.Content, input.NodeID)
	if node == nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "input field not found"})
	}

	mode, _ := node.Attrs["mode"].(string)
	if mode == "" {
		// Legacy inputEmbed nodes predate the "mode" attr and were always
		// regex-graded. Fall back to regex to preserve their behavior; new
		// nodes always serialize an explicit mode.
		mode = "regex"
	}
	pattern, _ := node.Attrs["pattern"].(string)
	caseInsensitive, _ := node.Attrs["caseInsensitive"].(bool)
	scoreF, _ := node.Attrs["score"].(float64)
	nodeScore := int(scoreF)

	var passed, graded bool
	score := 0
	switch mode {
	case "manual":
		// No auto-grade: stored as pending until an instructor grades it.
		graded = false
	case "regex":
		passed = s.matchValue(pattern, input.Value, caseInsensitive)
		graded = true
	default: // "exact": literal comparison (regex metachars are not special)
		passed = exactMatch(pattern, input.Value)
		graded = true
	}
	if passed {
		score = nodeScore
	}

	if err := s.repo.Create(ctx, &repositories.CreateInputSubmissionPayload{
		UserID:             input.UserID,
		NodeID:             input.NodeID,
		DocumentMaterialID: input.DocumentMaterialID,
		LabID:              input.LabID,
		SectionID:          input.SectionID,
		Value:              input.Value,
		Passed:             passed,
		Score:              score,
		Graded:             graded,
	}); err != nil {
		return nil, err
	}

	return &SubmitInputAnswerResult{Passed: passed, Score: score, Graded: graded}, nil
}

// GradeManualInput sets an instructor-assigned score on a submission, clamping
// it to [0, node score] and marking the submission graded.
func (s *inputSubmissionService) GradeManualInput(ctx context.Context, input *GradeInputInput) error {
	sub, err := s.repo.GetByID(ctx, input.SubmissionID)
	if err != nil {
		return err
	}
	if sub == nil {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "input submission not found"})
	}

	maxScore := s.nodeScore(ctx, sub.DocumentMaterialID, sub.NodeID)
	score := input.Score
	if score < 0 {
		score = 0
	}
	if maxScore > 0 && score > maxScore {
		score = maxScore
	}

	return s.repo.Grade(ctx, sub.ID, score, score > 0)
}

// nodeScore looks up the configured max score of an input embed node. Returns 0
// when the document or node cannot be resolved.
func (s *inputSubmissionService) nodeScore(ctx context.Context, documentMaterialID, nodeID string) int {
	doc, err := s.docRepo.GetByID(ctx, documentMaterialID)
	if err != nil || doc.Content == nil {
		return 0
	}
	var root inputTiptapNode
	if err := json.Unmarshal([]byte(*doc.Content), &root); err != nil {
		return 0
	}
	node := findInputEmbed(root.Content, nodeID)
	if node == nil {
		return 0
	}
	scoreF, _ := node.Attrs["score"].(float64)
	return int(scoreF)
}

// exactMatch reports whether the submitted value equals the expected value.
// Surrounding whitespace is trimmed so accidental spaces don't fail an answer.
func exactMatch(expected, value string) bool {
	return strings.TrimSpace(expected) == strings.TrimSpace(value)
}

// matchValue compiles the pattern and reports whether the entire value matches.
// A failed compile is treated as a non-match (logged, not fatal).
func (s *inputSubmissionService) matchValue(pattern, value string, caseInsensitive bool) bool {
	if caseInsensitive && !strings.HasPrefix(pattern, "(?i)") {
		pattern = "(?i)" + pattern
	}

	// Anchor so the full submitted value must match the pattern.
	anchored := "^(?:" + pattern + ")$"

	re, err := regexp.Compile(anchored)
	if err != nil {
		s.logger.Warnw("input submission regex failed to compile", "pattern", pattern, "error", err)
		return false
	}
	return re.MatchString(value)
}

func (s *inputSubmissionService) GetMyLatestResult(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*InputResult, error) {
	sub, err := s.repo.GetLatest(ctx, userID, nodeID, documentMaterialID, labID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return &InputResult{Submitted: false, Passed: false, Score: 0, Value: ""}, nil
	}
	return &InputResult{
		Submitted: true,
		Passed:    sub.Passed,
		Score:     sub.Score,
		Graded:    sub.Graded,
		Value:     sub.Value,
	}, nil
}

func (s *inputSubmissionService) ListByMaterial(ctx context.Context, documentMaterialID string) ([]InputSubmissionResult, error) {
	subs, err := s.repo.ListByMaterial(ctx, documentMaterialID)
	if err != nil {
		return nil, err
	}
	results := make([]InputSubmissionResult, 0, len(subs))
	for _, sub := range subs {
		results = append(results, InputSubmissionResult{
			ID:        sub.ID,
			UserID:    sub.UserID,
			NodeID:    sub.NodeID,
			Value:     sub.Value,
			Passed:    sub.Passed,
			Score:     sub.Score,
			Graded:    sub.Graded,
			CreatedAt: sub.CreatedAt,
		})
	}
	return results, nil
}
