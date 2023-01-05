#!/bin/bash
until pg_isready -h "localhost" -p "25432"
do
  sleep 2;
done

PGPASSWORD=postgres psql -U postgres -d postgres -h localhost -p 25432 <<EOF
CREATE DATABASE taco;
EOF

PGPASSWORD=postgres psql -U postgres -d taco -h localhost -p 25432 <<EOF
CREATE EXTENSION IF NOT EXISTS postgis;
EOF

atlas schema apply --exclude "topology*" --auto-approve -u "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable" --file .
