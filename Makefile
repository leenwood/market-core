.PHONY: build run migrate lint test

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

run:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate -dir=migrations

lint:
	golangci-lint run ./...

test:
	go test -race -count=1 ./...

test-integration:
	go test -race -count=1 -timeout=120s -tags=integration ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down
