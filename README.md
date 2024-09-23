# Go API Template

## Prerequisites

- [Install Go](https://go.dev/doc/install)
- [Install Docker](https://docs.docker.com/get-started/)

## Running the app locally

### Install the packages

```sh
go mod download
```

### Start the database

```sh
docker compose up --detach postgres
```

### Run the database migrations

```sh
docker compose run --rm dbmate migrate
```

### Start the app

```sh
go run cmd/server/main.go
```

## Running tests

```sh
ginkgo -v -r ./...
```

## Building the production image

```sh
docker build . --tag go-api-template:latest
```
