# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY inventory-service/go.mod inventory-service/go.sum ./inventory-service/
COPY proto/go.mod proto/go.sum ./proto/

WORKDIR /src/inventory-service
RUN go mod download

WORKDIR /src
COPY inventory-service ./inventory-service
COPY proto ./proto

WORKDIR /src/inventory-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/inventory-service ./cmd/inventory-service

FROM alpine:3.20 AS runner
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/inventory-service /usr/local/bin/inventory-service
COPY inventory-service/internal/migrations ./internal/migrations

USER app
EXPOSE 50052

CMD ["inventory-service"]
