package pubsub

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

type redisPubSub struct {
	c  *redis.Client
	mu sync.Mutex
}

func NewRedis(connStr string, password ...string) (PubSub, error) {
	if len(password) > 0 && password[0] != "" {
		opt := &redis.Options{
			Addr:     connStr,
			Password: password[0],
		}
		c := redis.NewClient(opt)
		return &redisPubSub{c: c}, nil
	}

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

func (r *redisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	sub := r.c.Subscribe(ctx, channel)

	payloadChan := make(chan []byte)

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := sub.Close()
				if err != nil {
					fmt.Println("Error on close redis subscribe :", err)
				}
				close(payloadChan)
				return
			case msg := <-sub.Channel():
				payloadChan <- []byte(msg.Payload)
			}
		}
	}()

	return payloadChan, nil
}

func (r *redisPubSub) UnSubscribe(ctx context.Context, channel string) error {
	return fmt.Errorf("No need to call UnSubscribe because it already unsubscribe on Subscribe method")
}
