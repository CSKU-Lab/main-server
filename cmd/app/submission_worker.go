package main

import (
	"context"
	"log"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

func startSubmissionWorker(ctx context.Context, db *sqlx.DB, config *configs.Config) {
	conn, err := pgx.Connect(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalln("Cannot ping to db")
	}

	_, err = conn.Exec(ctx, "LISTEN code_submissions_outbox_insert")
	if err != nil {
		log.Fatalln("Cannot subscribe code_submissions_outbox")
	}

	log.Println("Waiting for notifications")

	for {
		noti, err := conn.WaitForNotification(ctx)
		if err != nil {
			log.Printf("there is error receive notification")
		}

		log.Printf("Receive notification : ", noti.Payload)
	}

	// _, err := queue.NewRabbitMQ(config.RBMQ_SERVER_URL)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

}
