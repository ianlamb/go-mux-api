# go-mux-api

Learning exercise using Go, Mux, Postgres and Docker to create a basic CRUD API.

## Resources

- https://semaphoreci.com/community/tutorials/building-and-testing-a-rest-api-in-go-with-gorilla-mux-and-postgresql
- https://blog.logrocket.com/how-to-build-a-restful-api-with-docker-postgresql-and-go-chi/

## DB Setup

Install go migrate (Go 1.16+):

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Setup initial migration (don't need to run this again, but can use as reference for future migrations):

```bash
migrate create -ext sql -dir db/migrations -seq create_items_table
```

Run the migrations:

```bash
export POSTGRESQL_URL="postgres://risky:rainy2@localhost:5432/ror2?sslmode=disable"
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```

## Run Locally

```bash
docker compose up --build
```
