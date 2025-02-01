#!/bin/sh
set -e

# Wait for PostgreSQL
until /usr/bin/pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER; do
  echo "Waiting for PostgresSQL..."
  sleep 2
done

# Run migrations
migrate -path ./migrations -database "$DB_DSN" up

# Start main application
exec "$@"
