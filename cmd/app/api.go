package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/jmoiron/sqlx"
)

func startApiServer(ctx context.Context, db *sqlx.DB, config *configs.Config) {
	app, cleanup, err := initializeApp(ctx, config, db)
	if err != nil {
		log.Fatal("Failed to initialize app: ", err)
	}
	defer cleanup()

	port := fmt.Sprintf(":%v", config.Port)

	go func() {
		<-ctx.Done()

		log.Println("Received shutdown signal, shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	err = app.Listen(port)
	if err != nil {
		log.Fatal("Error starting server on Port ", port, ": ", err)
	}

	log.Println("Server stopped")
}
