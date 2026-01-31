package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/CSKU-Lab/main-server/configs"
)

func main() {
	config := configs.NewConfig()
	db := configs.NewDB(config)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mode := flag.String("mode", "all", "api | worker | all")
	flag.Parse()

	if *mode == "worker" || *mode == "all" {
		go startSubmissionWorker(ctx, db, config)
	}

	if *mode == "api" || *mode == "all" {
		startApiServer(ctx, db, config)
	}
}
