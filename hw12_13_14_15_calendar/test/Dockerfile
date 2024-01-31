FROM golang:1.21 as build

ENV CODE_DIR /go/src

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ENV CGO_ENABLED=0

CMD go test -v -tags integration ./test