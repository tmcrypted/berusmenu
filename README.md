# Berusmenu MVP

Berusmenu is a service that helps PVZ owners in Russia urgently find a worker for a shift when there is a risk of downtime and fines.

This workspace is a monorepo for the MVP. All microservices live in this single repository, maintaining strict architectural boundaries and database isolation.

## Структура монорепозитория

Текущий набор директорий (микросервисов и общих компонентов):

- `api-gateway`
- `auth-service`
- `inventory-service`
- `matching-service`
- `communication-service`
- `notification-service`
- `proto`
- `frontend`
- `infrastructure`

Все сервисы и фронтенд находятся в одном Git-репозитории.

## Продуктовый MVP

Роли:

- `owner` — владелец ПВЗ
- `worker` — исполнитель смены
- `moderator` — модератор ПВЗ

Ключевой путь:

1. Пользователь регистрируется по email и паролю.
2. Email подтверждается кодом.
3. Owner создает ПВЗ и отправляет на модерацию.
4. Moderator одобряет или отклоняет ПВЗ.
5. Owner создает vacancy только для approved ПВЗ.
6. Worker откликается на vacancy.
7. Сразу после response создается conversation.
8. Owner выбирает worker, создается assignment.
9. После `completed` или `cancelled` обе стороны могут оставить review.
10. Под review можно оставлять replies без вложенности глубже одного уровня.

## Ключевые архитектурные решения

- Одна PostgreSQL база, но отдельная схема на сервис.
- FK только внутри схем сервиса.
- Публичный API идет через `api-gateway`.
- Внутренние вызовы между backend-сервисами идут по gRPC.
- Kafka в MVP используется только для notification events.
- Redis в MVP используется для verification/rate limiting/presence.
- `notification-service` отделен от доменных сервисов и не имеет своей БД в MVP.
- Email provider в MVP: Mailganer.
- Chat создается сразу после `response`.
- Chat становится read-only после `assignment.completed` и `assignment.cancelled`.

## Notification Flow

Kafka topics in MVP:

- `auth.email_verification_requested`
- `matching.response_created`

Notification delivery channel:

- email only

Delivery выполняет `notification-service`, который читает события из Kafka и отправляет письма через Mailganer.

## Redis Usage

Redis используется для:

- `email_verify:{user_id}` -> verification code, TTL 10 minutes
- `email_verify_resend:{user_id}` -> resend cooldown, TTL 60 seconds
- `email_verify_attempts:{user_id}` -> 3 failed attempts -> lockout for 1 hour
- rate limiting
- online/offline presence
- typing status
- active websocket connections

Повторная отправка verification code пересылает тот же действующий код, пока он не истек и нет lockout.

## Unverified Users

При регистрации пользователь получает JWT, но пока email не подтвержден, backend разрешает только whitelist endpoints:

- `logout`
- `resend verification code`
- `verify email`
- `me` is not available until email verification succeeds

Все остальные действия запрещены на backend, а frontend перенаправляет пользователя на экран подтверждения email.

## Documentation Map

- [AGENTS.md](C:/Users/tmcrypted/Desktop/berusmenu/AGENTS.md)
- [Architecture](C:/Users/tmcrypted/Desktop/berusmenu/docs/architecture.md)
- [System Design](C:/Users/tmcrypted/Desktop/berusmenu/docs/system-design.md)
- [Sequence Diagrams](C:/Users/tmcrypted/Desktop/berusmenu/docs/sequence-diagrams.md)
- [State Transitions](C:/Users/tmcrypted/Desktop/berusmenu/docs/state-transitions.md)
- [Permissions](C:/Users/tmcrypted/Desktop/berusmenu/docs/permissions.md)
- [Interfaces](C:/Users/tmcrypted/Desktop/berusmenu/docs/interfaces.md)
- [Events And Cache](C:/Users/tmcrypted/Desktop/berusmenu/docs/events-and-cache.md)
- [Runtime Topology](C:/Users/tmcrypted/Desktop/berusmenu/docs/runtime-topology.md)
- [Development Plan](C:/Users/tmcrypted/Desktop/berusmenu/docs/development-plan.md)

## Что дальше

Следующие инженерные шаги:

1. Freeze gRPC and REST contracts.
2. Подготовить goose migrations по схемам.
3. Реализовать `notification-service`.
4. Реализовать `auth-service`.
5. Реализовать `inventory-service`.
6. Реализовать `matching-service`.
7. Реализовать `communication-service`.
8. Доработать `api-gateway`.
9. Реализовать frontend flows.
