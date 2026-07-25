#!/bin/sh
set -eu

MIGRATIONS_PATH="${MIGRATIONS_PATH:-/app/migrations}"

echo "running database migrations from ${MIGRATIONS_PATH}"
migrate -path "${MIGRATIONS_PATH}" -database "${DATABASE_URL}" up

echo "starting api"
exec /app/api
