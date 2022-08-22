.ONESHELL:
SHELL = /bin/bash

.PHONY: docker_build_local 

docker_build_base:
	docker build --tag taco_base:dev . -f ./docker/Dockerfile.server.base

docker_build_local: docker_build_base
	docker build --tag taco:dev . -f ./docker/Dockerfile.server.local

docker_migration_build_local:
	docker build --tag taco_migration:dev . -f ./docker/Dockerfile.migration.base
