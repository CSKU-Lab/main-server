package registrables

import "testing"

func TestEvaluateTypingScorePreservesTwoDecimalPlacesWhenScaled(t *testing.T) {
	score := evaluateTypingScore(20.0, 3.0)
	if score != 76.67 {
		t.Fatalf("evaluateTypingScore() = %v, want 76.67", score)
	}

	storedScore := int(score * typingScoreScale)
	if storedScore != 7667 {
		t.Fatalf("scaled score = %d, want 7667", storedScore)
	}
}
