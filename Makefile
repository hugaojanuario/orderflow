# OrderFlow — alvos de desenvolvimento local
#
# migrate-up/migrate-down usam o CLI do golang-migrate via `go run`,
# então não é preciso instalar nada além do Go.
# DATABASE_URL precisa estar exportada no ambiente (ou use make DATABASE_URL=...).

MIGRATE := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

.PHONY: run-api run-worker run-web test lint migrate-up migrate-down seed

run-api:
	cd api && go run ./cmd/api

run-worker:
	cd worker && go run ./cmd/worker

run-web:
	cd web && npm run dev

test:
	cd api && go test ./...
	cd worker && go test ./...

lint:
	cd api && golangci-lint run ./...
	cd worker && golangci-lint run ./...
	cd web && npm run lint

migrate-up:
	$(MIGRATE) -path api/db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path api/db/migrations -database "$(DATABASE_URL)" down 1

seed:
	cd api && go run ./cmd/api --seed
