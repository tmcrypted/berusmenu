# Infrastructure

Локальная инфраструктура Berusmenu для foundation-этапа.

## Что входит

- `docker-compose.yml`
- `.env.example`
- контейнерные слоты для всех app repos
- базовые runtime dependencies:
  - PostgreSQL
  - Redis
  - Kafka

## Что поднимается по умолчанию

Без дополнительных profiles поднимаются:

- `postgres`
- `redis`
- `kafka`
- `api-gateway`
- `frontend`

## Что пока оставлено в future profile

Пока доменные сервисы и notification worker не реализованы кодом, для них уже зафиксированы build contexts и контейнерные слоты:

- `auth-service`
- `inventory-service`
- `matching-service`
- `communication-service`
- `notification-service`

Запуск с ними:

```bash
docker compose --env-file .env.example --profile future up --build
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
2. Поднять `api-gateway`
3. Поднять `frontend`
4. Позже добавить future profile сервисы

## Notes

- `api-gateway` уже ориентирован на будущие gRPC адреса сервисов
- Kafka в MVP зарезервирована только под notification events
- Redis в MVP нужен для verification, rate limiting и realtime ephemeral state
