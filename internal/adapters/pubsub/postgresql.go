package pubsub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type postgresPubSub struct {
	conn        *pgx.Conn
	dataBaseURL string
	logger      *zap.SugaredLogger
}

func NewPostgres(ctx context.Context, logger *zap.SugaredLogger, dataBaseURL string) (PubSub, func() error, error) {
	conn, err := pgx.Connect(ctx, dataBaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, nil, errors.New("Cannot ping to db")
	}

	close := func() error {
		return conn.Close(ctx)
	}

	return &postgresPubSub{
		conn:        conn,
		dataBaseURL: dataBaseURL,
		logger:      logger,
	}, close, nil
}

func (p *postgresPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	_, err := p.conn.Exec(ctx, fmt.Sprintf("LISTEN \"%s\"", channel))
	if err != nil {
		return nil, fmt.Errorf("Cannot subscribe %s", channel)
	}

	payloadChan := make(chan []byte)

	go func() {
		defer close(payloadChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			noti, err := p.conn.WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				p.logger.Errorw("postgres pubsub subscribe error, reconnecting", "error", err)
				if !p.reconnect(ctx, channel) {
					// ctx cancelled during reconnect; stop the goroutine.
					return
				}
				continue
			}

			payloadChan <- []byte(noti.Payload)
		}
	}()

	return payloadChan, nil

}

// reconnect re-establishes the connection and re-issues LISTEN, retrying with
// capped exponential backoff until it succeeds. It returns false only when ctx
// is cancelled. A single-shot attempt is not enough: on a postgres restart the
// server rejects connections for a few seconds ("the database system is
// starting up"), so one failed reconnect would silently kill the listener and
// leave the worker depending on the 5-minute reconciliation sweep alone.
func (p *postgresPubSub) reconnect(ctx context.Context, channel string) bool {
	_ = p.conn.Close(ctx)

	const (
		baseDelay = 1 * time.Second
		maxDelay  = 30 * time.Second
	)
	delay := baseDelay

	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(delay):
		}

		newConn, connErr := pgx.Connect(ctx, p.dataBaseURL)
		if connErr != nil {
			p.logger.Errorw("postgres pubsub reconnect failed, retrying", "error", connErr, "next_retry", delay)
			delay = min(delay*2, maxDelay)
			continue
		}

		if _, listenErr := newConn.Exec(ctx, fmt.Sprintf("LISTEN \"%s\"", channel)); listenErr != nil {
			p.logger.Errorw("postgres pubsub re-LISTEN failed, retrying", "error", listenErr, "next_retry", delay)
			_ = newConn.Close(ctx)
			delay = min(delay*2, maxDelay)
			continue
		}

		p.conn = newConn
		p.logger.Infow("postgres pubsub reconnected and re-LISTENing", "channel", channel)
		return true
	}
}

func (p *postgresPubSub) UnSubscribe(ctx context.Context, channel string) error {
	_, err := p.conn.Exec(ctx, fmt.Sprintf("UNLISTEN \"%s\"", channel))
	if err != nil {
		return fmt.Errorf("Cannot unsubscribe %s", channel)
	}
	return nil
}

func (p *postgresPubSub) Publish(ctx context.Context, channel string, message string) error {
	return fmt.Errorf("Not Implemented")
}
