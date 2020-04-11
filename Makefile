ENV_LOCAL_FILE := .env.local
ENV_LOCAL = $(shell cat $(ENV_LOCAL_FILE))

.PHONY: run-db-local
run-db-local:
	$(ENV_LOCAL) docker-compose -f docker/docker-compose.deps.base.yml -f docker/docker-compose.deps.local.yml -p local up -d
