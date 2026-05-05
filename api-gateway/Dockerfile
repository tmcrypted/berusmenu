# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
COPY proto/go.mod proto/go.sum ./proto/

WORKDIR /src/api-gateway
RUN go mod download

WORKDIR /src
COPY api-gateway ./api-gateway
COPY proto ./proto

WORKDIR /src/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.20 AS runner
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway

USER app
EXPOSE 8080

CMD ["api-gateway"]
