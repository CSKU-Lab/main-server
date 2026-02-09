package pubsub

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type postgresPubSub struct {
	conn *pgx.Conn
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
		conn: conn,
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
				fmt.Println("Error in Postgres PubSub Subscribe:", err)
				return
			}

			payloadChan <- []byte(noti.Payload)
		}
	}()

	return payloadChan, nil

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
