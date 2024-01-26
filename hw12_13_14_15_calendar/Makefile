BIN_CALENDAR := "./bin/calendar"
BIN_CALENDAR_SCHEDULER := "./bin/calendar_scheduler"
BIN_CALENDAR_SENDER := "./bin/calendar_sender"

DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

RABBITMQNAME := rabbitmq

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

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
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

stop_rabbitmq:
	docker stop $(RABBITMQNAME)

run_rabbitmq:
	docker run -d --rm --name $(RABBITMQNAME) -e RABBITMQ_DEFAULT_USER=root -e RABBITMQ_DEFAULT_PASS=secretkey -p 15672:15672 -p 5672:5672 rabbitmq:3-management

run_postgres:
	docker run -d --rm --name pg -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secretkey -e PGDATA=/var/lib/postgresql/data/pgdata -v psqldata:/var/lib/postgresql/data -p 5432:5432 postgres:14

create_db:
	docker exec -it pg createdb --username=root --owner=root calendar

drop_db:
	docker exec -it pg dropdb calendar

migrate_up:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" up

migrate_down:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=calendar sslmode=disable" down

.PHONY: build build-calendar build-calendar-scheduler build-calendar-sender
.PHONY: run-calendar run-calendar-scheduler run-calendar-sender
.PHONY: build-img run-img version test lint run_postgres createdb dropdb
