package routes

import (
	"context"
	"encoding/json"

	"github.com/CSKU-Lab/queue"
)

type OGImageEvent struct {
	Type     string  `json:"type"`
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle *string `json:"subtitle,omitempty"`
	Tag      *string `json:"tag,omitempty"`
}

func publishOGEvent(q queue.Queue, event OGImageEvent) {
	go func() {
		payload, err := json.Marshal(event)
		if err != nil {
			return
		}
		_ = q.Publish(context.Background(), "", "og_image", &queue.Derivery{
			Body: payload,
		})
	}()
}
