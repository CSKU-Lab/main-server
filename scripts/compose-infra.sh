#!/bin/sh

export COMPOSE_PROJECT_NAME=main-server-infra

doppler run -- docker compose -f docker/docker-compose.infra.yaml "$@"
