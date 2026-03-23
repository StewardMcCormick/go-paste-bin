.PHONY: gen-mocks
gen-mocks:
	@mockery

.PHONY: test
test: gen-mocks
	@go test ./...

.PHONY: test-unit
test-unit: gen-mocks
	@go test $(shell go list ./... | grep -v ./test)

.PHONY: test-with-cover
test-with-cover: gen-mocks
	@go test -cover ./...

.PHONY: test-integrations
test-integrations:
	@go test -v -count=1 ./test/...

.PHONY: build-local
build-local:
	@go build -o ./bin/app ./cmd/app/main.go
	@echo " ✅ Built local binary file"

.PHONY: build-docker-single
build-docker-single:
	@docker build -t paste-bin-api:latest .
	@echo " ✅ Built docker image"

.PHONY: build-docker-compose
build-docker-compose:
	@docker-compose build
	@echo " ✅ Built docker compose"

.PHONY: run-local
run-local: build-local
	@./bin/app

.PHONY: run-docker-single
run-docker-single: build-docker-single
	@docker run -p 8080:8080 --env-file .env.docker-single --name paste-bin-api --rm paste-bin-api:latest

.PHONY: run-docker-compose
run-docker-compose: build-docker-compose
	@docker compose up

.PHONY: clean-local
clean-local:
	@rm -rf ./bin
	@echo "Cleaned local binary files"

.PHONY: clean-docker
clean-docker:
	docker rm paste-bin-api

.PHONY: help
help:
	@echo "Supported command with 'make':"
	@echo "- gen-mocks: generate mocks for unit-tests"
	@echo "- test: run all tests without coverage"
	@echo "- test-with-cover: run all tests with coverage"
	@echo "- test-unit: run all unit-tests"
	@echo "- test-integrations: run all integration-tests"
	@echo "- build-local: compile project in binary file 'app.exe' locate in ./bin"
	@echo "- build-docker-single: build docker image 'paste-bin-api:latest'"
	@echo "- build-docker-compose: build docker compose on docker-compose.yaml"
	@echo "- run-local: run binary file created by 'build-local'"
	@echo "- run-docker-single: run docker container 'paste-bin-api' on 8080 port with '.env.docker-single' env file and remove it after shutting down"
	@echo "- run-docker-compose: run built docker-compose"
	@echo "- clean-local: delete built binary file"
	@echo "- clean-docker: remove built docker container 'paste-bin-api' - !Return Error if container not exists!"
