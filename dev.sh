#!/usr/bin/env bash
# Local dev: port-forward cluster deps + load env + run main-server with hot reload (air).
# Env is scoped to this script's subshell — nothing leaks into your interactive shell.
# Ctrl-C tears down every port-forward.
#
#   ./dev.sh
#
set -euo pipefail

NS=cs-lab
kgs() { kubectl -n "$NS" get secret "$1" -o jsonpath="{.data.$2}" | base64 -d; }

# --- port-forward deps -------------------------------------------------------
pids=()
for pf in \
  postgres:5432:5432 \
  cache:6379:6379 \
  rabbitmq:5672:5672 \
  config-server:8081:8081 \
  task-server:50051:50051 \
  go-grader-master:50052:50052 \
  s3:9000:9000 \
  ; do
  svc=${pf%%:*}; ports=${pf#*:}
  kubectl -n "$NS" port-forward "svc/$svc" "$ports" >/dev/null 2>&1 &
  pids+=($!)
done
trap 'echo; echo "stopping port-forwards..."; kill "${pids[@]}" 2>/dev/null' EXIT
echo "port-forwarding: postgres cache rabbitmq config-server task-server go-grader-master s3"
sleep 2   # let tunnels establish before the app dials them

# --- env (subshell-scoped) ---------------------------------------------------
export DEV_MODE=true
export API_URL=http://localhost:8080
export PORT=8080

# cookie → localhost:3000 : empty domain = host-only cookie on localhost,
# shared across ports, so the web at :3000 receives it.
export COOKIE_DOMAIN=
export FRONTEND_URL=http://localhost:3000

# deps (via the port-forwards above)
export REDIS_ADDR=localhost:6379
export CONFIG_SERVER_URL=localhost:8081
export TASK_SERVER_URL=localhost:50051
export GRADER_SERVER_URL=localhost:50052

# S3 / MinIO (client inits at boot — endpoint must be set or the app exits)
export S3_ENDPOINT=localhost:9000
export S3_USE_SSL=false
export S3_BUCKET=cs-lab
export S3_FRONTEND_URL=http://localhost:9000

# secrets pulled live from the cluster (never written to disk)
export REDIS_PASSWORD="$(kgs cache-secrets REDIS_PASSWORD)"
export S3_ACCESS_KEY_ID="$(kgs main-server-api-secrets S3_ACCESS_KEY_ID)"
export S3_SECRET_ACCESS_KEY="$(kgs main-server-api-secrets S3_SECRET_ACCESS_KEY)"
export DATABASE_URL="postgresql://cs_pg_user:$(kgs main-server-api-secrets POSTGRES_PASSWORD)@localhost:5432/main-server?sslmode=disable"
export RBMQ_SERVER_URL="amqp://admin:$(kgs main-server-api-secrets RABBITMQ_PASSWORD)@localhost:5672/"
export JWT_SECRET="$(kgs main-server-api-secrets JWT_SECRET)"
export JWT_REFRESH_SECRET="$(kgs main-server-api-secrets JWT_REFRESH_SECRET)"
export INTERNAL_TOKEN="$(kgs main-server-api-secrets INTERNAL_TOKEN)"

# --- run with hot reload -----------------------------------------------------
# air inherits the exported env; rebuilds on .go changes. Uses .air.toml.
# NOT exec'd: keep this shell alive so the EXIT trap tears down port-forwards.
air
