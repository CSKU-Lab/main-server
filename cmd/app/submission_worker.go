package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

type outboxDeps struct {
	logger                   *zap.SugaredLogger
	submissionRepo           repositories.SubmissionRepository
	codeSubmissionRepo       repositories.CodeSubmissionRepository
	codeSubmissionOutboxRepo repositories.CodeSubmissionOutboxRepository
	rClient                  pubsub.PubSub
	q                        queue.Queue
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

	deps := &outboxDeps{
		logger:                   logger,
		submissionRepo:           submissionRepo,
		codeSubmissionRepo:       codeSubmissionRepo,
		codeSubmissionOutboxRepo: codeSubmissionOutboxRepo,
		rClient:                  rClient,
		q:                        q,
	}

	logger.Infoln("Submission Worker running...")

	unsentRecords, err := codeSubmissionOutboxRepo.GetUnsent(ctx, 100, 2*time.Minute)
	if err != nil {
		logger.Errorln("Failed to fetch unsent outbox records", "error", err)
	} else if len(unsentRecords) > 0 {
		logger.Infof("Found %d unsent outbox records, reprocessing...", len(unsentRecords))
		for _, record := range unsentRecords {
			go processOutboxRecord(ctx, deps, &notiPayload{
				ID:           record.ID,
				SubmissionID: record.SubmissionID,
				IsSent:       record.IsSent,
				Payload:      record.Payload,
			})
		}
	}

	go startReconciliationTicker(ctx, deps)

	msgs, err := sub.Subscribe(ctx, "code_submissions_outbox_insert")
	if err != nil {
		logger.Fatalln("Error while subscribing to code submissions outbox", "error", err)
	}

	for msg := range msgs {
		go func(m []byte) {
			var subPayload notiPayload
			err = json.Unmarshal(m, &subPayload)
			if err != nil {
				logger.Errorln("Cannot unmarshal code submission outbox message", "error", err)
				return
			}
			processOutboxRecord(ctx, deps, &subPayload)
		}(msg)
	}
}

func startReconciliationTicker(ctx context.Context, deps *outboxDeps) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deps.logger.Infoln("Reconciliation: scanning for unsent outbox records...")
			unsentRecords, err := deps.codeSubmissionOutboxRepo.GetUnsent(ctx, 100, 2*time.Minute)
			if err != nil {
				deps.logger.Errorln("Reconciliation: failed to fetch unsent records", "error", err)
				continue
			}
			if len(unsentRecords) == 0 {
				continue
			}
			deps.logger.Infof("Reconciliation: found %d unsent records, reprocessing...", len(unsentRecords))
			for _, record := range unsentRecords {
				go processOutboxRecord(ctx, deps, &notiPayload{
					ID:           record.ID,
					SubmissionID: record.SubmissionID,
					IsSent:       record.IsSent,
					Payload:      record.Payload,
				})
			}
		}
	}
}

func processOutboxRecord(ctx context.Context, deps *outboxDeps, subPayload *notiPayload) {
	qName, err := deps.q.CreateQueue(ctx, "grade_result-"+subPayload.ID, &queue.QueueOptions{
		AutoDelete: true,
	})
	if err != nil {
		deps.logger.Errorln("Cannot create grade result queue", "error", err)
		deps.codeSubmissionOutboxRepo.IncrementRetry(ctx, subPayload.ID)
		return
	}

	err = deps.q.Publish(ctx, "", "grade", &queue.Derivery{
		CorrelationID: subPayload.ID,
		ReplyTo:       qName,
		Body:          []byte(subPayload.Payload),
	})
	if err != nil {
		deps.logger.Errorln("Cannot publish grade request message", "error", err)
		deps.codeSubmissionOutboxRepo.IncrementRetry(ctx, subPayload.ID)
		return
	}

	marked, err := deps.codeSubmissionOutboxRepo.TryMarkSent(ctx, subPayload.ID)
	if err != nil {
		deps.logger.Errorln("Cannot mark code submission outbox as sent", "error", err)
	}
	if !marked {
		deps.logger.Warnf("Outbox record %s was already marked as sent by another instance, skipping result consumption", subPayload.ID)
		return
	}

	channel := fmt.Sprintf("submissions:update:%s", subPayload.SubmissionID)
	result := &models.GradeResult{}
	err = deps.q.Consume(ctx, qName, 1, true, func(derivery *queue.Derivery, exit chan struct{}) error {
		err := json.Unmarshal(derivery.Body, result)
		if err != nil {
			deps.logger.Errorln("Cannot unmarshal grade result message", "error", err)
			return err
		}

		deps.logger.Infof("Received grade result for submission_id %s", subPayload.SubmissionID)

		if result.Status != models.CODE_EXECUTION_QUEUED && result.Status != models.CODE_EXECUTION_RUNNING {
			exit <- struct{}{}
		}

		if result.Status == models.CODE_EXECUTION_QUEUED {
			status := models.QUEUED
			err := deps.submissionRepo.Update(ctx, &repositories.UpdateSubmissionRequest{
				ID:     subPayload.SubmissionID,
				Status: &status,
			})
			if err != nil {
				return err
			}

			err = deps.rClient.Publish(ctx, channel, string(models.QUEUED))
			if err != nil {
				return err
			}
		}

		if result.Status == models.CODE_EXECUTION_RUNNING {
			status := models.RUNNING
			err := deps.submissionRepo.Update(ctx, &repositories.UpdateSubmissionRequest{
				ID:     subPayload.SubmissionID,
				Status: &status,
			})
			if err != nil {
				return err
			}

			err = deps.rClient.Publish(ctx, channel, string(models.RUNNING))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		deps.logger.Errorln("Cannot consume grade result message", "error", err)
		return
	}

	err = deps.codeSubmissionRepo.Update(ctx, &repositories.UpdateCodeSubmissionPayload{
		SubmissionID:   subPayload.SubmissionID,
		Status:         string(result.Status),
		AvgWallTime:    result.AvgWallTime,
		AvgMemory:      result.AvgMemory,
		TestCaseGroups: result.TestCaseGroupResults,
	})
	if err != nil {
		deps.logger.Errorln("Cannot update code submission result", "error", err)
		return
	}

	var status models.SubmissionStatus
	if result.Status == models.CODE_EXECUTION_RUN_FAILED {
		status = models.FAILED
	} else {
		status = models.PASSED
	}

	autoScore := int(result.Score)
	err = deps.submissionRepo.Update(ctx, &repositories.UpdateSubmissionRequest{
		ID:        subPayload.SubmissionID,
		Status:    &status,
		AutoScore: &autoScore,
	})
	if err != nil {
		deps.logger.Errorln("Cannot update submission status", "error", err)
		return
	}

	err = deps.rClient.Publish(ctx, channel, string(status))
	if err != nil {
		deps.logger.Errorln("Cannot publish final submission status", "error", err)
		return
	}
}
