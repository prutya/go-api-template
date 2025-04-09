# Go API app Template

An opinionated DIY template for Golang JSON API applications

![Build status](https://github.com/prutya/go-api-template/actions/workflows/main.yml/badge.svg)

## What's in the box?

### App
- [x] Router ([chi](https://github.com/go-chi/chi))
- [x] Real IP middleware ([chi](https://pkg.go.dev/github.com/go-chi/chi/middleware#RealIP))
- [x] Structured logging middleware ([zap](https://github.com/uber-go/zap))
- [x] Request ID middleware
- [x] Error recovery middleware
- [x] CORS middleware ([chi](https://github.com/go-chi/cors))
- [x] Compatibility with standard library (net/http) middleware
- [x] Error handling
- [x] Healthcheck endpoint `/health`
- [x] Configurable Authentication via refresh token
- [x] Sign In endpoint – `POST /sessions`
- [x] Sign Out endpoint – `DELETE /sessions/current`
- [x] Get current User endpoint – `GET /users/current`

### Database
- [x] ORM ([bun](https://github.com/uptrace/bun))
- [x] Language-agnostic database migration toolkit ([dbmate](https://github.com/amacneil/dbmate))

### Quality control
- [x] Testing setup ([ginkgo](https://github.com/onsi/ginkgo))
- [x] Github Actions Test job
- [x] Github Actions Lint job ([golangci-lint](https://github.com/golangci/golangci-lint))

### Misc
- [x] Structured logger ([zap](https://github.com/uber-go/zap))
- [x] Configuration ([viper](https://github.com/spf13/viper))

### Development and deployment
- [x] Docker Compose setup for development
- [x] Multi-stage Dockerfile

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

### Seed the database

```sh
docker compose run --rm psql --echo-all --file /db/seed.sql
```

### Start the app

```sh
go run cmd/server/main.go
```

### Recreating the database
```sh
# Drop the database
docker compose run --rm dbmate drop

# Create the database and run migrations
docker compose run --rm dbmate up
```

## Building the production image

```sh
docker build . --tag my-counters-api:latest
```

## Running tests

```sh
docker compose run --build --rm test
```

## Running the linter
```sh
docker compose run --build --rm lint
```

## Building the production image

```sh
docker build . --tag go-api-template:latest
```
