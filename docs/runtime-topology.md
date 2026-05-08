# Runtime Topology

## Production shape

```mermaid
flowchart TB
    Internet --> Nginx["Nginx + SSL"]
    Nginx --> Frontend["Frontend"]
    Nginx --> Gateway["api-gateway"]
    Gateway --> Auth["auth-service"]
    Gateway --> Inventory["inventory-service"]
    Gateway --> Matching["matching-service"]
    Gateway --> Communication["communication-service"]

    Auth --> Postgres["PostgreSQL"]
    Inventory --> Postgres
    Matching --> Postgres
    Communication --> Postgres

    Auth --> Redis["Redis"]
    Gateway --> Redis
    Communication --> Redis

    Auth --> Kafka["Kafka"]
    Matching --> Kafka
    Kafka --> Notification["notification-service"]
    Notification --> Email["Email sender"]
```

## Notes

- `notification-service` is a background worker.
- `notification-service` has no own DB in MVP.
- WebSocket traffic terminates at gateway / communication edge.
- Browser cookies terminate at `api-gateway`; downstream services receive identity metadata, not cookies.
- Gateway allows credentialed CORS only for configured frontend origins.
- Postgres schemas are service-owned.
- Redis state is ephemeral.
- Kafka carries only notification events in MVP.
- Local development uses notification email stub logs; Mailganer is planned as production provider.

## Local Docker shape

Default local compose stack:

```mermaid
flowchart TB
    Browser["Browser / Insomnia"] --> Gateway["api-gateway :8080"]
    Frontend["frontend :3000"] --> Gateway

    Gateway --> Auth["auth-service :50051"]
    Gateway --> Inventory["inventory-service :50052"]

    Auth --> Postgres["postgres :5432"]
    Inventory --> Postgres
    Auth --> Redis["redis :6379"]
    Auth --> Kafka["redpanda/kafka :9092"]

    Kafka --> Notification["notification-service"]
```

Kafka listeners:

- containers use `kafka:9092`
- host `go run` services use `localhost:19092`

Local API-only run:

```bash
docker compose --env-file .env up -d postgres redis kafka auth-service inventory-service notification-service api-gateway
```

Local frontend origin defaults to `http://localhost:3000`; gateway CORS must include that origin so browser cookies work across `localhost:3000` -> `localhost:8080`.

Full local run with frontend:

```bash
docker compose --env-file .env up --build
```
