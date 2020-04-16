ENV_LOCAL_FILE := .env.local
ENV_LOCAL = $(shell cat $(ENV_LOCAL_FILE))
ENV_TEST_FILE := .env.test
ENV_TEST = $(shell cat $(ENV_TEST_FILE))


.PHONY:serve
serve:
	$(ENV_LOCAL) go run main.go

.PHONY:test
test:
	$(ENV_TEST) go test -v ./...

.PHONY: run-db-local
run-db-local:
	$(ENV_LOCAL) docker-compose -f docker/docker-compose.deps.base.yml -f docker/docker-compose.deps.local.yml -p local up -d
