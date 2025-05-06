##### Base #####

FROM golang:1.24.2-bookworm@sha256:79390b5e5af9ee6e7b1173ee3eac7fadf6751a545297672916b59bfa0ecf6f71 AS base



##### Dependencies #####

FROM base AS deps

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download



##### Testing #####

FROM deps AS test

# Install Ginkgo
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest



##### Production build #####

FROM deps AS build_production

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

# Build the server binary first, so that the build cache can be reused by the
# other builds. This is faster than building the binaries in parallel.
RUN go build -ldflags="-s -w" -o server ./cmd/server/main.go

# Build the worker binary
RUN go build -ldflags="-s -w" -o worker ./cmd/worker/main.go

# Create a non-root user
RUN echo "app:x:1000:1000:App:/:" > /etc_passwd



##### Final production image ####

FROM debian:bookworm-20250428-slim@sha256:4b50eb66f977b4062683ff434ef18ac191da862dbe966961bc11990cf5791a8d AS production

RUN apt update && apt install -y ca-certificates curl && apt clean

WORKDIR /app

COPY --from=build_production /app/server .
COPY --from=build_production /app/worker .

# Prepare the config file
RUN echo "{}" > app.json

# Copy the non-root user
COPY --from=build_production /etc_passwd /etc/passwd

USER app

CMD ["/app/server"]

EXPOSE 3333
