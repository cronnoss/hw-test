BIN_CALENDAR := "./bin/calendar"
BIN_CALENDAR_SCHEDULER := "./bin/calendar_scheduler"
BIN_CALENDAR_SENDER := "./bin/calendar_sender"

DOCKER_IMG1="calendar:develop"
DOCKER_IMG2="calendar_scheduler:develop"
DOCKER_IMG3="calendar_sender:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

RABBITMQNAME := rabbitmq

Red='\033[0;31m'
Green='\033[0;32m'
Color_Off='\033[0m'

help:
	@echo ${Red}"Please select a subcommand"${Color_Off}
	@echo ${Green}"make run-postgres"${Color_Off}" to run postgres"
	@echo ${Green}"make create-db"${Color_Off}" to create db"
	@echo ${Green}"make migrate-up"${Color_Off}" to migrate up"
	@echo ${Green}"make run-rabbitmq"${Color_Off}" to run rabbitmq"
	@echo
	@echo ${Green}"make build"${Color_Off}" to build applications"
	@echo ${Green}"make run-calendar"${Color_Off}" to run calendar"
	@echo ${Green}"make run-calendar-scheduler"${Color_Off}" to run scheduler"
	@echo ${Green}"make run-calendar-sender"${Color_Off}" to run sender"
	@echo
	@echo ${Green}"make generate"${Color_Off}" to generate stub-files from protobuf-files"
	@echo
	@echo ${Red}"Or use docker-compose:"
	@echo ${Green}"make up"${Color_Off}" to run docker-compose"
	@echo ${Green}"make down"${Color_Off}" to stop docker-compose"
	@echo ${Green}"make destroy"${Color_Off}" to stop docker-compose and remove volumes"
	@echo
	@echo ${Green}"make test"${Color_Off}" to run unit tests"
	@echo ${Green}"make integration-test"${Color_Off}" to run integration tests"

build: build-calendar build-calendar-scheduler build-calendar-sender

build-calendar:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar

run-calendar: build-calendar
	$(BIN_CALENDAR) -config ./configs/calendar_config.toml

build-calendar-scheduler:
	go build -v -o $(BIN_CALENDAR_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/scheduler

run-calendar-scheduler: build-calendar-scheduler
	$(BIN_CALENDAR_SCHEDULER) -config ./configs/scheduler_config.toml

build-calendar-sender:
	go build -v -o $(BIN_CALENDAR_SENDER) -ldflags "$(LDFLAGS)" ./cmd/sender

run-calendar-sender: build-calendar-sender
	$(BIN_CALENDAR_SENDER) -config ./configs/sender_config.toml

build-img-calendar:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG1) \
		-f build/calendar/Dockerfile .

build-img-scheduler:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG2) \
		-f build/scheduler/Dockerfile .

build-img-sender:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG3) \
		-f build/sender/Dockerfile .

build-img: build-img-calendar build-img-scheduler build-img-sender

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_CALENDAR) version

test:
	go clean -testcache;
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

lint: install-lint-deps
	golangci-lint run ./...

install-grpc-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate: install-grpc-deps
	mkdir -p pkg/event_service_v1
	protoc 	--proto_path=api --go_out=pkg/event_service_v1 --go_opt=paths=source_relative \
			--go-grpc_out=pkg/event_service_v1  --go-grpc_opt=paths=source_relative \
			EventService.proto

stop-rabbitmq:
	docker stop $(RABBITMQNAME)

run-rabbitmq:
	docker run -d --rm --name $(RABBITMQNAME) -e RABBITMQ_DEFAULT_USER=root -e RABBITMQ_DEFAULT_PASS=secretkey -p 15672:15672 -p 5672:5672 rabbitmq:3-management

run-postgres:
	docker run -d --rm --name pg -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secretkey -e PGDATA=/var/lib/postgresql/data/pgdata -v psqldata:/var/lib/postgresql/data -p 5432:5432 postgres:latest

create-db:
	docker exec -it pg createdb --username=root --owner=root calendar

drop-db:
	docker exec -it pg dropdb calendar

migrate-up:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" down

migrate-status:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" status

migrate-reset:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" reset

up:
	@docker-compose -f ./deployments/docker-compose.yaml up -d

down:
	@docker-compose -f ./deployments/docker-compose.yaml down

destroy:
	@docker-compose -f ./deployments/docker-compose.yaml down -v

integration-test:
	set -e ;\
	docker-compose -f ./deployments/docker-compose.yaml up -d ;\
	test_status_code=0 ;\
	go test -v ./test/integration_test.go || test_status_code=$$? ;\
	docker-compose -f ./deployments/docker-compose.yaml down ;\
	docker-compose -f ./deployments/docker-compose.yaml down -v ;\
	exit $$test_status_code ;

.PHONY: build build-calendar build-calendar-scheduler build-calendar-sender
.PHONY: run-calendar run-calendar-scheduler run-calendar-sender
.PHONY: build-img run-img version test lint run-postgres create-db drop-db run-rabbitmq stop-rabbitmq
.PHONY: migrate-up migrate-down generate install-grpc-deps install-lint-deps help