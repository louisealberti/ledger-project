# Variables - Adjust these based on your .env
DB_CONTAINER_NAME=ledger-postgres
DB_NAME=ledger_db
DB_USER=postgres

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: Start the application (compiles main and seed)
run:
	go run ./cmd/ledger

## test: Run all unit and integration tests
test:
	go test -v ./...

## db-up: Start the database container
db-up:
	docker-compose up -d

## db-shell: Enter the database shell inside Docker
db-shell:
	docker exec -it $(DB_CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

## db-clean: Remove the database container and volumes (Fresh start)
db-clean:
	docker-compose down -v

## tidy: Clean up go.mod and download dependencies
tidy:
	go mod tidy