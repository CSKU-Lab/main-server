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

// RegradeResult summarizes a document material regrade pass.
type RegradeResult struct {
	// Regraded is the number of auto-mode submissions re-evaluated.
	Regraded int `json:"regraded"`
	// Skipped is the number of submissions left untouched (manual mode or the
	// node no longer exists in the document).
	Skipped int `json:"skipped"`
}

type InputSubmissionService interface {
	SubmitInputAnswer(ctx context.Context, input *SubmitInputAnswerInput) (*SubmitInputAnswerResult, error)
	GetMyLatestResult(ctx context.Context, userID, nodeID, documentMaterialID, labID string) (*InputResult, error)
	ListByMaterial(ctx context.Context, documentMaterialID string) ([]InputSubmissionResult, error)
	// GradeManualInput sets an instructor-assigned score on a manual-mode submission.
	GradeManualInput(ctx context.Context, input *GradeInputInput) error
	// RegradeMaterial re-evaluates auto-mode input submissions against the
	// document's current node config, without students resubmitting.
	RegradeMaterial(ctx context.Context, documentMaterialID string) (*RegradeResult, error)
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

	_, passed, graded, score := s.evalInputNode(node, input.Value)

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

// evalInputNode grades a submitted value against an input embed node's current
// config. It returns the resolved mode so callers can special-case manual
// grading (which carries no auto score). An empty mode resolves to "regex" to
// preserve legacy nodes that predate the mode attr.
func (s *inputSubmissionService) evalInputNode(node *inputTiptapNode, value string) (mode string, passed, graded bool, score int) {
	mode, _ = node.Attrs["mode"].(string)
	if mode == "" {
		mode = "regex"
	}
	pattern, _ := node.Attrs["pattern"].(string)
	caseInsensitive, _ := node.Attrs["caseInsensitive"].(bool)
	scoreF, _ := node.Attrs["score"].(float64)
	nodeScore := int(scoreF)

	switch mode {
	case "manual":
		// No auto-grade: stays pending until an instructor grades it.
		graded = false
	case "regex":
		passed = s.matchValue(pattern, value, caseInsensitive)
		graded = true
	default: // "exact": literal comparison (regex metachars are not special)
		passed = exactMatch(pattern, value)
		graded = true
	}
	if passed {
		score = nodeScore
	}
	return mode, passed, graded, score
}

// RegradeMaterial re-evaluates the latest input submissions of a document
// material against the document's current node config, without students
// resubmitting. Manual-mode submissions keep their instructor grade; nodes that
// no longer exist in the document are skipped.
func (s *inputSubmissionService) RegradeMaterial(ctx context.Context, documentMaterialID string) (*RegradeResult, error) {
	doc, err := s.docRepo.GetByID(ctx, documentMaterialID)
	if err != nil {
		return nil, err
	}
	if doc.Content == nil {
		return &RegradeResult{}, nil
	}

	var root inputTiptapNode
	if err := json.Unmarshal([]byte(*doc.Content), &root); err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "invalid document content"})
	}

	subs, err := s.repo.ListByMaterial(ctx, documentMaterialID)
	if err != nil {
		return nil, err
	}

	res := &RegradeResult{}
	for i := range subs {
		sub := &subs[i]
		node := findInputEmbed(root.Content, sub.NodeID)
		if node == nil {
			res.Skipped++
			continue
		}
		mode, passed, _, score := s.evalInputNode(node, sub.Value)
		if mode == "manual" {
			// Preserve the instructor-assigned grade.
			res.Skipped++
			continue
		}
		if err := s.repo.Grade(ctx, sub.ID, score, passed); err != nil {
			return nil, err
		}
		res.Regraded++
	}
	return res, nil
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
