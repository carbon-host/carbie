include .env

build:
	go build -o ./bin/bot/${BINARY} ./cmd/bot/main.go

install: build
	go install ./cmd/bot

start:
	./bin/bot/${BINARY} ${ARGS}

restart: build start