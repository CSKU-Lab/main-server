#!/bin/sh

export COMPOSE_PROJECT_NAME=main-server-infra

doppler run --command="docker compose -f docker/docker-compose.infra.yaml up -d"
