#!/bin/sh
set -e
echo "==> Running DB migrations..."
go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 \
  -dir /migrations \
  postgres \
  "$DATABASE_URL" \
  up
echo "==> Migrations done!"
