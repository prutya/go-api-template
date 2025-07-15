# Go API app Template

An opinionated DIY template for Golang JSON API applications

![Build status](https://github.com/prutya/go-api-template/actions/workflows/main.yml/badge.svg)

## What's in the box?

### App
- [x] Router ([chi](https://github.com/go-chi/chi))
- [x] Real IP middleware ([chi](https://pkg.go.dev/github.com/go-chi/chi/middleware#RealIP))
- [x] Request ID middleware
- [x] Error recovery middleware
- [x] CORS middleware ([chi](https://github.com/go-chi/cors))
- [x] Compatibility with standard library (net/http) middleware
- [x] Error handling
- [x] Secure Configurable Authentication (based on Refresh Tokens)
- [x] Transactional Emails via [Scaleway](https://www.scaleway.com/)
- [x] CAPTCHA via [Cloudflare Turnstile](https://www.cloudflare.com/application-services/products/turnstile/)

### Database
- [x] ORM ([bun](https://github.com/uptrace/bun))
- [x] Language-agnostic database migration toolkit ([dbmate](https://github.com/amacneil/dbmate))

### Background jobs processing
- [x] Background jobs processing setup via [Asynq](https://github.com/hibiken/asynq)

### Quality control
- [x] Testing setup ([ginkgo](https://github.com/onsi/ginkgo))
- [x] Github Actions Test job
- [x] Github Actions Lint job ([golangci-lint](https://github.com/golangci/golangci-lint))

### Misc
- [x] Structured logger ([slog](https://go.dev/blog/slog))
- [x] Configuration ([viper](https://github.com/spf13/viper))

### Development and deployment
- [x] Docker Compose setup for development
- [x] Multi-stage Dockerfile

## Prerequisites

- [Install Go](https://go.dev/doc/install)
- [Install Docker](https://docs.docker.com/get-started/)

## Running the app locally

### 1. Install the packages

```sh
go mod download
```

### 2. Start the database and Redis servers

```sh
docker compose up postgres redis
```

### 3. Run the database migrations

```sh
docker compose run --rm dbmate migrate
```

### 4. Seed the database

```sh
docker compose run --rm psql --echo-all --file /db/seed.sql
```

### 5. Start the app

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

## Running the background jobs processor locally

### 1. Set up the database
Make sure that steps 1-4 from **Running the app locally** are completed

### 2. Start the worker
```sh
go run cmd/worker/main.go
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

Both server and worker binaries will be in the same image

```sh
docker build . --tag go-api-template:latest
```
