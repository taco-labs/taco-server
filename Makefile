.ONESHELL:
SHELL = /bin/bash

.PHONY: docker_build_local 

docker_server_build_base:
	docker build --tag taco:dev . -f ./docker/Dockerfile.server.base

docker_server_build_local: docker_server_build_base
	docker build --tag taco_local:dev . -f ./docker/Dockerfile.server.local

docker_migration_build_base:
	docker build --tag taco_migration:dev . -f ./docker/Dockerfile.migration.base

docker_build_local: docker_server_build_local docker_migration_build_base 

run_local: docker_build_local
	docker-compose -f ./docker/docker-compose.yaml up -d

stop_local: 
	docker-compose -f ./docker/docker-compose.yaml down
