package pubsub

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

type redisPubSub struct {
	c  *redis.Client
	mu sync.Mutex
}

func NewRedis(connStr string, password ...string) (PubSub, error) {
	pw := ""
	if len(password) > 0 {
		pw = password[0]
	}

	var c *redis.Client
	if strings.Contains(connStr, "://") {
		opt, err := redis.ParseURL(connStr)
		if err != nil {
			return nil, err
		}
		c = redis.NewClient(opt)
	} else {
		c = redis.NewClient(&redis.Options{
			Addr:     connStr,
			Password: pw,
		})
	}

	return &redisPubSub{c: c}, nil
}

func (r *redisPubSub) Publish(ctx context.Context, channel string, message string) error {
	return r.c.Publish(ctx, channel, message).Err()
}

func (r *redisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	sub := r.c.Subscribe(ctx, channel)

	// Block until the server confirms the subscription is active. Without this,
	// r.c.Subscribe returns before SUBSCRIBE is acked, so a caller that re-checks
	// state right after Subscribe could still miss a message published in the
	// window. Draining the confirmation here makes the returned channel live.
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()
		return nil, err
	}

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
