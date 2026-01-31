package main

import (
	"context"
	"log"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/queue"
	"github.com/jmoiron/sqlx"
)

func startSubmissionWorker(ctx context.Context, db *sqlx.DB, config *configs.Config) {
	_, err := queue.NewRabbitMQ(config.RBMQ_SERVER_URL)
	if err != nil {
		log.Fatalln(err)
	}

}
