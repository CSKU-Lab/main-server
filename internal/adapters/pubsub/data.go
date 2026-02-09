package pubsub

import (
	"context"
	"errors"
)

var (
	Exit error = errors.New("exit subscribe")
)

type handler func(message []byte) error

type PubSub interface {
	Publish(ctx context.Context, channel string, message string) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
	UnSubscribe(ctx context.Context, channel string) error
}
