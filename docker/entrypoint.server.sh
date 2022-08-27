#!/bin/bash

until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT"
do
  sleep 2;
done

exec "$@"
