# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY notification-service/go.mod notification-service/go.sum ./notification-service/
COPY proto/go.mod proto/go.sum ./proto/

WORKDIR /src/notification-service
RUN go mod download

WORKDIR /src
COPY notification-service ./notification-service
COPY proto ./proto

WORKDIR /src/notification-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/notification-service ./cmd/notification-service

FROM alpine:3.20 AS runner
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/notification-service /usr/local/bin/notification-service
COPY notification-service/internal/migrations ./internal/migrations

USER app

CMD ["notification-service"]
