# Infrastructure

Локальная инфраструктура Berusmenu для foundation-этапа.

## Что входит

- `docker-compose.yml`
- `.env.example`
- контейнерные слоты для всех app workspaces
- базовые runtime dependencies:
  - PostgreSQL
  - Redis
  - Kafka

## Что поднимается по умолчанию

Без дополнительных profiles поднимаются:

- `postgres`
- `redis`
- `kafka`
- `auth-service`
- `notification-service`
- `api-gateway`
- `frontend`

## Что пока оставлено в future profile

Пока доменные сервисы не реализованы кодом, для них уже зафиксированы build contexts и контейнерные слоты:

- `inventory-service`
- `matching-service`
- `communication-service`

Запуск с ними:

```bash
docker compose --env-file .env --profile future up --build
```

## Local API run

Для проверки только backend API без frontend:

```bash
docker compose --env-file .env up -d postgres redis kafka auth-service notification-service api-gateway
```

Если app-сервисы запускаются локально через `go run`, а инфраструктура остается в Docker, Kafka broker должен быть:

```bash
KAFKA_BROKERS=localhost:19092
```

Для сервисов внутри compose используется:

```bash
KAFKA_BROKERS=kafka:9092
```

Для полного локального запуска с frontend:

```bash
docker compose --env-file .env up --build
```

## Migration strategy

Migrations не централизуются здесь.

Source of truth:

- `auth-service/internal/migrations`
- `inventory-service/internal/migrations`
- `matching-service/internal/migrations`
- `communication-service/internal/migrations`

`infrastructure` только фиксирует порядок и runtime story.

## Local bootstrap order

1. Поднять `postgres`, `redis`, `kafka`
2. Поднять `auth-service` и `notification-service`
3. Поднять `api-gateway`
4. Опционально поднять `frontend`
5. Позже добавить future profile сервисы

## Notes

- `api-gateway` уже ориентирован на будущие gRPC адреса сервисов
- Local web auth uses credentialed CORS from `http://localhost:3000` to `http://localhost:8080`
- Production must set `COOKIE_SECURE=true` and configure HTTPS before browser auth cookies are used
- Kafka inside compose is `kafka:9092`; host-side `go run` services should use `KAFKA_BROKERS=localhost:19092`
- Kafka в MVP зарезервирована только под notification events
- Redpanda healthcheck проверяет `rpk cluster health` и допускает пробелы в строке `Healthy: true`
- Redis в MVP нужен для verification, rate limiting и realtime ephemeral state
