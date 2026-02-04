package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

type redisPubSub struct {
	c  *redis.Client
	mu sync.Mutex
}

func NewRedis(connStr string) (PubSub, error) {
	opt, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, err
	}

	c := redis.NewClient(opt)

	return &redisPubSub{
		c: c,
	}, nil
}

func (r *redisPubSub) Publish(ctx context.Context, channel string, message string) error {
	return r.c.Publish(ctx, channel, message).Err()
}

func (r *redisPubSub) Subscribe(ctx context.Context, channel string, handler handler) error {
	sub := r.c.Subscribe(ctx, channel)
	defer sub.Close()

	errChan := make(chan error)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case msg := <-sub.Channel():
			go func() {
				err := handler([]byte(msg.Payload))
				if err != nil {
					if errors.Is(err, Exit) {
						errChan <- nil
					}
					errChan <- err
				}
			}()
		}
	}
}

func (r *redisPubSub) UnSubscribe(ctx context.Context, channel string) error {
	return fmt.Errorf("No need to call UnSubscribe because it already unsubscribe on Subscribe method")
}
