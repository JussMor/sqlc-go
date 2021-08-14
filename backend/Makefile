network:
	docker network create bank-network

postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrate:
	migrate -path postgres/migrations \
		-database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" up

rollback:
	migrate -path postgres/migrations \
		-database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"  down

drop:
	migrate -path postgres/migrations \
		-database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"  drop

migration:
	@read -p "Enter migration name: " name; \
		migrate create -ext sql -dir postgres/migrations $$name

sqlc:
	docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc generate

server:
	go run main.go


.PHONY: network  postgres createdb dropdb migrate rollback drop migration sqlc server