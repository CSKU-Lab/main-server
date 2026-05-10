#!/bin/sh
set -e

COMPOSE="docker compose -f docker/docker-compose.test.yaml -p main-server-test"
TEST_DB_URL="postgres://cs_pg_user:cs_pg_password@localhost:5433/main-server?sslmode=disable"

teardown() {
  $COMPOSE down --volumes --remove-orphans
}
trap teardown EXIT

$COMPOSE up -d --wait

atlas schema apply \
  --url "$TEST_DB_URL" \
  --to "file://atlas/schema.hcl" \
  --auto-approve

PGHOST=localhost \
PGPORT=5433 \
PGUSER=cs_pg_user \
PGPASSWORD=cs_pg_password \
PGDATABASE=main-server \
  go test -tags integration -v -count=1 "$@"
