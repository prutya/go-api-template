##### Base #####

FROM golang:1.24.2-bookworm@sha256:00eccd446e023d3cd9566c25a6e6a02b90db3e1e0bbe26a48fc29cd96e800901 AS base



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

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -tags netgo -o main ./cmd/server/main.go

# Create a non-root user
RUN echo "app:x:1000:1000:App:/:" > /etc_passwd



##### Final production image ####

FROM debian:bookworm-20250407-slim@sha256:b1211f6d19afd012477bd34fdcabb6b663d680e0f4b0537da6e6b0fd057a3ec3 AS production

RUN apt update && apt install -y ca-certificates curl && apt clean

WORKDIR /app

COPY --from=build_production /app/main .

# Prepare the config file
RUN echo "{}" > app.json

# Copy the non-root user
COPY --from=build_production /etc_passwd /etc/passwd

USER app

CMD ["/app/main"]

EXPOSE 3333
