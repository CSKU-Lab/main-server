package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/internal/logging"
)

func main() {
	config := configs.NewConfig()
	db := configs.NewDB(config)

	logger, loggerCleanup, err := logging.New(os.Getenv("ENV"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := loggerCleanup(); err != nil {
			logger.Warnw("failed to flush logger", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mode := flag.String("mode", "all", "api | worker | all")
	flag.Parse()

	if *mode == "worker" || *mode == "all" {
		go startSubmissionWorker(ctx, logger, db, config)
	}

	if *mode == "api" || *mode == "all" {
		startApiServer(ctx, db, config)
	}
}
