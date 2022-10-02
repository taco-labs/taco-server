.ONESHELL:
SHELL = /bin/bash

.PHONY: docker_build_local 

docker_server_build_base:
	docker build --tag taco:dev . -f ./docker/Dockerfile.server.base

docker_server_build_local: docker_server_build_base
	docker build --tag taco_local:dev . -f ./docker/Dockerfile.server.local

docker_server_build_prod: docker_server_build_base
	docker build --tag 069049357473.dkr.ecr.ap-northeast-2.amazonaws.com/taco/taco-backend:latest . -f ./docker/Dockerfile.server.prod

docker_migration_build_base:
	docker build --tag taco_migration:dev . -f ./docker/Dockerfile.migration.base

docker_build_local: docker_server_build_local docker_migration_build_base 

docker_build: docker_server_build_base docker_migration_build_base

run_local: docker_build_local
	docker-compose -f ./docker/docker-compose.yaml up -d
	./migrations/migration.sh

stop_local: 
	docker-compose -f ./docker/docker-compose.yaml down
