ENV_LOCAL_FILE := env.local
ENV_LOCAL = $(shell cat $(ENV_LOCAL_FILE))
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

.PHONY:test-with-coverage
test-with-coverage:
	$(ENV_TEST) $(ENV_SECRET_EXAMPLE) go test -v ./... -count=1 -covermode=atomic -coverprofile=profile.out

.PHONY: integration-test
integration-test:
	$(ENV_TEST) $(ENV_SECRET) go test -v ./... -tags integration -count=1

.PHONY: run-db-local
run-db-local:
	$(ENV_LOCAL) docker-compose -f docker/docker-compose.deps.base.yml -f docker/docker-compose.deps.local.yml -p local up -d

.PHONY:generate
generate:
	go generate ./...

.PHONY: tools
tools:
	cat tools/tools.go | grep "_" | awk -F'"' '{print $$2}' | xargs -tI % go install %

