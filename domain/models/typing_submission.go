package models

type TypingSubmission struct {
	RawWPM      float64 `json:"raw_wpm"`
	AdjustedWPM float64 `json:"adjusted_wpm"`
	ErrorRate   float64 `json:"error_rate"`
	Duration    float64 `json:"duration"`
}
