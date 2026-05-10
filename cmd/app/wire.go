//go:build wireinject

package main

import (
	"context"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/providers"
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func initializeApp(ctx context.Context, cfg *configs.Config, db *sqlx.DB, logger *zap.SugaredLogger) (*fiber.App, func(), error) {
	wire.Build(
		providers.RepositorySet,
		providers.ServiceSet,
		providers.RegistrySet,
		providers.ExternalSet,
		providers.NewFiberApp,
		middlewares.NewErrorHandlerMiddleware,
	)
	return nil, nil, nil
}
