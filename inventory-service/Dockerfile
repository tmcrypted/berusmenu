# syntax=docker/dockerfile:1

FROM alpine:3.20
WORKDIR /app

COPY README.md ./README.md
COPY AGENTS.md ./AGENTS.md
COPY internal/migrations ./internal/migrations

CMD ["sh", "-c", "echo 'inventory-service foundation image placeholder'; tail -f /dev/null"]
