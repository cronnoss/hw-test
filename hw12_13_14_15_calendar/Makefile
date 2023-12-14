BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.toml

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

.PHONY: build run build-img run-img version test lint run_postgres createdb dropdb
