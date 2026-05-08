# Infrastructure

Local Berusmenu infrastructure lives in `infrastructure/docker-compose.yml`.

## Default Stack

Without additional profiles, compose starts:

- `postgres`
- `redis`
- `kafka` / Redpanda
- `auth-service`
- `inventory-service`
- `notification-service`
- `api-gateway`
- `frontend`

`matching-service` and `communication-service` are kept under the `future` profile because their runtime implementation is still planned.

## Ports

Values are loaded from `infrastructure/.env`.

Typical local ports:

- Postgres host port: `54320`
- Redis host port: `6379`
- Kafka internal listener: `9092`
- Kafka host listener: `19092`
- Auth gRPC: `50051`
- Inventory gRPC: `50052`
- API Gateway HTTP: `8080`
- Frontend: `3000`

If a port is occupied by a local service, change the value in `.env` instead of editing `docker-compose.yml`.

## Run Infrastructure Only

Use this when Go services are started manually from the host:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\infrastructure
docker compose --env-file .env up -d postgres redis kafka
```

Host-side addresses for `go run`:

```powershell
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:KAFKA_BROKERS="localhost:19092"
```

Docker service names work only inside the compose network:

```text
postgres:5432
redis:6379
kafka:9092
```

## Run Backend In Docker

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\infrastructure
docker compose --env-file .env up -d postgres redis kafka auth-service inventory-service notification-service api-gateway
```

## Run Full Local Stack

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\infrastructure
docker compose --env-file .env up --build
```

## Run Future Profile

`matching-service` and `communication-service` are placeholders under the `future` profile:

```powershell
docker compose --env-file .env --profile future up --build
```

This profile is not part of the normal MVP runtime yet.

## Migrations

Migrations are owned by services, not by `infrastructure`.

Current implemented migration sources:

- `auth-service/internal/migrations`
- `inventory-service/internal/migrations`

Planned migration sources:

- `matching-service/internal/migrations`
- `communication-service/internal/migrations`

When a Go service starts, it applies its own migrations from `MIGRATIONS_DIR`, defaulting to:

```text
internal/migrations
```

For host-side `go run`, start from the service root:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\auth-service
go run .\cmd\auth-service
```

This keeps `internal/migrations` resolvable.

## Notes

- Local web auth uses credentialed CORS from `http://localhost:3000` to `http://localhost:8080`.
- Production must set `COOKIE_SECURE=true` and serve over HTTPS.
- Redpanda exposes `kafka:9092` inside compose and `localhost:19092` to the host.
- `notification-service` is a background worker and has no PostgreSQL schema in MVP.
- Redis is used for verification, rate limiting and future realtime ephemeral state.
