#!/bin/bash

until pg_isready -h "$TACO_DATABASE_HOST" -p "$TACO_DATABASE_PORT"
do
  sleep 2;
done

exec "$@"
