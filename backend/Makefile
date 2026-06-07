.PHONY: run build test vet fmt tidy migrate-up migrate-down

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race

vet:
	go vet ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir db/migrations postgres "$(DATABASE_URL)" down
