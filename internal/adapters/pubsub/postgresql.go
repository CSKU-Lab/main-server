package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func NewPostgres(ctx context.Context, logger *zap.SugaredLogger, dataBaseURL string) (*pgx.Conn, func() error, error) {
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

	return conn, close, nil
}

func Listen[T any](ctx context.Context, conn *pgx.Conn, channel string, handler func(payload T) error) error {
	notiChan := make(chan T, 100)

	var eg errgroup.Group
	eg.Go(func() error {
		_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel))
		if err != nil {
			return fmt.Errorf("Cannot subscribe %s", channel)
		}

		for {
			noti, err := conn.WaitForNotification(ctx)
			if err != nil {
				return errors.New("there is error receive notification")
			}

			var payload T
			err = json.Unmarshal([]byte(noti.Payload), &payload)
			if err != nil {
				return err
			}

			notiChan <- payload
		}
	})

	eg.Go(func() error {
		for noti := range notiChan {
			eg.Go(func() error {
				return handler(noti)
			})
		}
		return nil
	})

	return eg.Wait()
}
