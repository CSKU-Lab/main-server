package models

type TypingMaterial struct {
	Content     string  `json:"content"`
	MinAdjWPM   float64 `json:"min_adj_wpm"`
	MinAccuracy float64 `json:"min_accuracy"`
}
