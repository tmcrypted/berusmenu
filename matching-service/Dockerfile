# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
ARG GOPROXY="https://proxy.golang.org|https://goproxy.io|https://goproxy.cn|direct"
ENV GOPROXY=${GOPROXY}
WORKDIR /src

COPY matching-service/go.mod matching-service/go.sum ./matching-service/
COPY proto/go.mod proto/go.sum ./proto/

WORKDIR /src/matching-service
RUN go mod download

WORKDIR /src
COPY matching-service ./matching-service
COPY proto ./proto

WORKDIR /src/matching-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/matching-service ./cmd/matching-service

FROM alpine:3.20 AS runner
WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/matching-service /usr/local/bin/matching-service
COPY matching-service/internal/migrations ./internal/migrations

USER app
EXPOSE 50053

CMD ["matching-service"]
