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

atlas schema apply --auto-approve -u "postgresql://postgres:postgres@localhost:25432/taco?sslmode=disable" -f migrations
