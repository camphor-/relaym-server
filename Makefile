ENV_LOCAL_FILE := env.local
ENV_LOCAL = $(shell cat $(ENV_LOCAL_FILE))
ENV_DEV_FILE := env.dev
ENV_DEV = $(shell cat $(ENV_DEV_FILE))
ENV_TEST_FILE := env.test
ENV_TEST = $(shell cat $(ENV_TEST_FILE))

ENV_SECRET_FILE = env.secret
ENV_SECRET = $(shell cat $(ENV_SECRET_FILE))
ENV_SECRET_EXAMPLE_FILE = env.secret.example
ENV_SECRET_EXAMPLE = $(shell cat $(ENV_SECRET_EXAMPLE_FILE))

.PHONY:serve
serve:
	$(ENV_LOCAL) $(ENV_SECRET) go run main.go

.PHONY:test
test:
	$(ENV_TEST) $(ENV_SECRET_EXAMPLE) go test -v ./... -count=1

.PHONY: run-db-local
run-db-local:
	$(ENV_LOCAL) docker-compose -f docker/docker-compose.deps.base.yml -f docker/docker-compose.deps.local.yml -p local up -d

.PHONY:generate
generate:
	go generate ./...

.PHONY: deploy-dev
deploy-dev:
	ENV=dev
	$(ENV_DEV) ENV_FILE=$(ENV_DEV_FILE) docker-compose -f docker/docker-compose.base.yml -f docker/docker-compose.$(ENV).yml -p $(ENV) stop
	$(ENV_DEV) ENV_FILE=$(ENV_DEV_FILE) docker-compose -f docker/docker-compose.base.yml -f docker/docker-compose.$(ENV).yml -p $(ENV) up -d