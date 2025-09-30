#!/bin/sh

export COMPOSE_PROJECT_NAME=main-server-infra

docker compose -f docker/docker-compose.infra.yaml --env-file .env "$@"
