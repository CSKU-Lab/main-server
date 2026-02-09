package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/queue"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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

	rClient, err := pubsub.NewRedis(config.REDIS_SERVER_URL)
	if err != nil {
		logger.Fatalln(err)
	}

	q, err := queue.NewRabbitMQ(config.RBMQ_SERVER_URL)
	if err != nil {
		logger.Fatalln(err)
	}

	sub, close, err := pubsub.NewPostgres(ctx, logger, config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}
	defer close()

	logger.Infoln("Submission Worker running...")

	msgs, err := sub.Subscribe(ctx, "code_submissions_outbox_insert")
	if err != nil {
		logger.Fatalln("Error while subscribing to code submissions outbox", "error", err)
	}

	for msg := range msgs {
		go func() {
			var subPayload notiPayload
			err = json.Unmarshal(msg, &subPayload)
			if err != nil {
				logger.Errorln("Cannot unmarshal code submission outbox message", "error", err)
			}

			qName, err := q.CreateQueue(ctx, "grade_result-"+subPayload.ID, &queue.QueueOptions{
				AutoDelete: true,
			})
			if err != nil {
				logger.Errorln("Cannot create grade result queue", "error", err)
			}

			err = q.Publish(ctx, "", "grade", &queue.Derivery{
				CorrelationID: subPayload.ID,
				ReplyTo:       qName,
				Body:          []byte(subPayload.Payload),
			})
			if err != nil {
				logger.Errorln("Cannot publish grade request message", "error", err)
			}

			err = codeSubmissionOutboxRepo.Update(ctx, subPayload.ID, true)
			if err != nil {
				logger.Errorln("Cannot mark code submission outbox as sent", "error", err)
			}

			channel := fmt.Sprintf("submissions:update:%s", subPayload.SubmissionID)
			result := &models.GradeResult{}
			err = q.Consume(ctx, qName, 1, true, func(derivery *queue.Derivery, exit chan struct{}) error {
				err := json.Unmarshal(derivery.Body, result)
				if err != nil {
					logger.Errorln("Cannot unmarshal grade result message", "error", err)
					return err
				}

				logger.Infof("Received grade result for submission_id %s", subPayload.SubmissionID)

				if result.Status != models.CODE_EXECUTION_QUEUED && result.Status != models.CODE_EXECUTION_RUNNING {
					exit <- struct{}{}
				}

				if result.Status == models.CODE_EXECUTION_QUEUED {
					err := submissionRepo.Update(ctx, subPayload.SubmissionID, models.QUEUED)
					if err != nil {
						return err
					}

					err = rClient.Publish(ctx, channel, string(models.QUEUED))
					if err != nil {
						return err
					}
				}

				if result.Status == models.CODE_EXECUTION_RUNNING {
					err := submissionRepo.Update(ctx, subPayload.SubmissionID, models.RUNNING)
					if err != nil {
						return err
					}

					err = rClient.Publish(ctx, channel, string(models.RUNNING))
					if err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				logger.Errorln("Cannot consume grade result message", "error", err)
			}

			err = codeSubmissionRepo.Update(ctx, &repositories.UpdateCodeSubmissionPayload{
				SubmissionID:   subPayload.SubmissionID,
				Status:         string(result.Status),
				AvgWallTime:    result.AvgWallTime,
				AvgMemory:      result.AvgMemory,
				TestCaseGroups: result.TestCaseGroupResults,
			})
			if err != nil {
				logger.Errorln("Cannot update code submission result", "error", err)
				return
			}

			var status models.SubmissionStatus
			if result.Status == models.CODE_EXECUTION_RUN_FAILED {
				status = models.FAILED
			} else {
				status = models.PASSED
			}

			err = submissionRepo.Update(ctx, subPayload.SubmissionID, status)
			if err != nil {
				logger.Errorln("Cannot update submission status", "error", err)
				return
			}

			err = rClient.Publish(ctx, channel, string(status))
			if err != nil {
				logger.Errorln("Cannot publish final submission status", "error", err)
				return
			}
		}()
	}
}
