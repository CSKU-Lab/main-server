package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/internal/logging"
	cskuotel "github.com/CSKU-Lab/otel"
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

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		otelShutdown, err := cskuotel.Init(context.Background())
		if err != nil {
			logger.Warnw("tracing unavailable", "error", err)
		} else {
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := otelShutdown(shutdownCtx); err != nil {
					logger.Warnw("tracer shutdown error", "error", err)
				}
			}()
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mode := flag.String("mode", "all", "api | worker | all")
	flag.Parse()

	if *mode == "worker" || *mode == "all" {
		go startSubmissionWorker(ctx, logger, db, config)
	}

	if *mode == "api" || *mode == "all" {
		startApiServer(ctx, db, config, logger)
	}

	if *mode == "worker" {
		<-ctx.Done()
	}
}
