FROM golang:1.13.8

ENV BOT_TOKEN ""

WORKDIR /app
COPY ./app/go.mod ./
COPY ./app/go.sum ./

RUN go mod download

COPY ./app/main.go ./
COPY ./translate.db ./

RUN go build -o /app/mtbot

ENTRYPOINT ["/app/mtbot"]