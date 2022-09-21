#!/bin/bash
set -e

PGPASSWORD="$POSTGRES_PASSWORD" psql -v ON_ERROR_STOP=1 --host "$POSTGRES_HOST" --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOF
  CREATE USER aurora PASSWORD 'aurora';
  CREATE DATABASE aurora;
  GRANT ALL PRIVILEGES ON DATABASE aurora TO aurora;
EOF
PGPASSWORD="aurora" psql -v ON_ERROR_STOP=1 --host "$POSTGRES_HOST" --username aurora --dbname aurora < ./.github/scripts/init.txt
