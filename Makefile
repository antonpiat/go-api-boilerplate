.PHONY: run test test-integration lint build migrate-up migrate-down migrate-create docker-up docker-down tidy

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/gin-api?sslmode=disable
MIGRATE ?= migrate

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test -race -count=1 ./...

test-integration:
	INTEGRATION=1 go test -race -count=1 ./internal/server -run TestAuthAndUserFlow

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

migrate-up:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=add_foo"; exit 1)
	$(MIGRATE) create -ext sql -dir migrations -seq $(name)

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
