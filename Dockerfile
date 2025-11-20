##### Base #####

FROM golang:1.25.4-trixie AS base



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
# other builds
# TODO: Use the release build command when the code is stable
RUN go build -tags=debug -o server ./cmd/server/main.go
# RUN go build -ldflags="-s -w" -o server ./cmd/server/main.go

# Build the worker and scheduler binaries in parallel
# NOTE: This is of course not ideal, in the real world you would probably want
# to build separate images for worker and scheduler.
# TODO: Use the release build command when the code is stable
RUN go build -tags=debug -o worker ./cmd/worker/main.go & \
    go build -tags=debug -o scheduler ./cmd/scheduler/main.go & \
    wait
# RUN go build -ldflags="-s -w" -o worker ./cmd/worker/main.go & \
#     go build -ldflags="-s -w" -o scheduler ./cmd/scheduler/main.go & \
#     wait

# Create a non-root user
RUN echo "app:x:1000:1000:App:/:" > /etc_passwd



##### Final production image ####

FROM gcr.io/distroless/static-debian12:nonroot AS production

WORKDIR /app

COPY --from=build_production /app/server .
COPY --from=build_production /app/worker .
COPY --from=build_production /app/scheduler .
COPY ./config ./config

CMD ["/app/server"]

EXPOSE 3333
