package main

import (
	"context"
	"fmt"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func startApiServer(ctx context.Context, db *sqlx.DB, config *configs.Config, logger *zap.SugaredLogger) {
	app, cleanup, err := initializeApp(ctx, config, db, logger)
	if err != nil {
		logger.Fatalw("failed to initialize app", "error", err)
	}
	defer cleanup()

	port := fmt.Sprintf(":%v", config.Port)

	go func() {
		<-ctx.Done()

		logger.Infow("received shutdown signal, shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Errorw("error during server shutdown", "error", err)
		}
	}()

	err = app.Listen(port)
	if err != nil {
		logger.Fatalw("error starting server", "port", port, "error", err)
	}

	logger.Infow("server stopped")
}
