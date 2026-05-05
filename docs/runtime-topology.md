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
    Notification --> Mailganer["Mailganer"]
```

## Notes

- `notification-service` is a background worker.
- `notification-service` has no own DB in MVP.
- WebSocket traffic terminates at gateway / communication edge.
- Postgres schemas are service-owned.
- Redis state is ephemeral.
- Kafka carries only notification events in MVP.
