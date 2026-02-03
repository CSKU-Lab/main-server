package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/queue"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type notiPayload struct {
	ID           string `json:"id"`
	SubmissionID string `json:"submission_id"`
	IsSent       bool   `json:"is_sent"`
	Payload      string `json:"payload"`
}

func startSubmissionWorker(ctx context.Context, logger *zap.SugaredLogger, db *sqlx.DB, config *configs.Config) {
	submissionRepo := sqlxAdapter.NewSubmissionRepository(db)
	codeSubmissionRepo := sqlxAdapter.NewCodeSubmission(db)
	codeSubmissionOutboxRepo := sqlxAdapter.NewCodeSubmissionOutboxRepository(db)

	q, err := queue.NewRabbitMQ(config.RBMQ_SERVER_URL)
	if err != nil {
		logger.Fatalln(err)
	}

	listener, err := newPostgresListener(ctx, logger, config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	err = listener.Listen(ctx, func(payload *notiPayload) error {
		qName, err := q.CreateQueue(ctx, "grade_result-"+payload.ID)
		if err != nil {
			return err
		}

		err = q.Publish(ctx, "", "grade", &queue.Derivery{
			CorrelationID: payload.ID,
			ReplyTo:       qName,
			Body:          []byte(payload.Payload),
		})
		if err != nil {
			return err
		}

		err = codeSubmissionOutboxRepo.Update(ctx, payload.ID, true)
		if err != nil {
			return err
		}

		result := &models.GradeResult{}
		err = q.Consume(ctx, qName, 1, func(derivery *queue.Derivery, exit chan struct{}) error {
			err := json.Unmarshal(derivery.Body, result)
			if err != nil {
				logger.Errorln("Cannot unmarshal grade result message", "error", err)
				return err
			}

			logger.Infof("Received grade result for submission_id %s", payload.SubmissionID)

			if result.Status != models.CODE_EXECUTION_QUEUED && result.Status != models.CODE_EXECUTION_RUNNING {
				exit <- struct{}{}
			}

			if result.Status == models.CODE_EXECUTION_QUEUED {
				err := submissionRepo.Update(ctx, payload.SubmissionID, models.QUEUED)
				if err != nil {
					return err
				}
			}

			if result.Status == models.CODE_EXECUTION_RUNNING {
				err := submissionRepo.Update(ctx, payload.SubmissionID, models.RUNNING)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			logger.Fatalln(err)
		}

		err = codeSubmissionRepo.Update(ctx, &repositories.UpdateCodeSubmissionPayload{
			SubmissionID:   payload.SubmissionID,
			Status:         string(result.Status),
			AvgWallTime:    result.AvgWallTime,
			AvgMemory:      result.AvgMemory,
			TestCaseGroups: result.TestCaseGroupResults,
		})
		if err != nil {
			return err
		}

		var status models.SubmissionStatus
		if result.Status == models.CODE_EXECUTION_RUN_FAILED {
			status = models.FAILED
		} else {
			status = models.PASSED
		}

		err = submissionRepo.Update(ctx, payload.SubmissionID, status)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Fatalln(err)
	}

}

type postgresNotifier struct {
	conn   *pgx.Conn
	logger *zap.SugaredLogger
}

func newPostgresListener(ctx context.Context, logger *zap.SugaredLogger, dataBaseURL string) (*postgresNotifier, error) {
	conn, err := pgx.Connect(ctx, dataBaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, errors.New("Cannot ping to db")
	}

	return &postgresNotifier{
		conn:   conn,
		logger: logger,
	}, nil
}

func (p *postgresNotifier) Listen(ctx context.Context, handler func(payload *notiPayload) error) error {
	defer p.conn.Close(ctx)

	notiChan := make(chan *notiPayload, 100)

	var eg errgroup.Group
	eg.Go(func() error {
		_, err := p.conn.Exec(ctx, "LISTEN code_submissions_outbox_insert")
		if err != nil {
			return errors.New("Cannot subscribe code_submissions_outbox")
		}

		p.logger.Info("Waiting for notifications")

		for {
			noti, err := p.conn.WaitForNotification(ctx)
			if err != nil {
				return errors.New("there is error receive notification")
			}

			var payload notiPayload
			err = json.Unmarshal([]byte(noti.Payload), &payload)
			if err != nil {
				return err
			}

			notiChan <- &payload
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
