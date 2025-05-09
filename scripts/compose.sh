#!/bin/sh

export COMPOSE_PROJECT_NAME=main-server

docker compose -f docker/docker-compose.dev.yaml --env-file .env "$@"
