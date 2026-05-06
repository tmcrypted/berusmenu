# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY auth-service/go.mod auth-service/go.sum ./auth-service/
COPY proto/go.mod proto/go.sum ./proto/

WORKDIR /src/auth-service
RUN go mod download

WORKDIR /src
COPY auth-service ./auth-service
COPY proto ./proto

WORKDIR /src/auth-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/auth-service ./cmd/auth-service

FROM alpine:3.20 AS runner
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/auth-service /usr/local/bin/auth-service
COPY auth-service/internal/migrations ./internal/migrations

USER app
EXPOSE 50051

CMD ["auth-service"]
