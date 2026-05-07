# Berusmenu MVP

Berusmenu is a service that helps PVZ owners in Russia urgently find a worker for a shift when there is a risk of downtime and fines.

This workspace is a monorepo for the MVP. All microservices live in this single repository, maintaining strict architectural boundaries and database isolation.

Canonical Go import path:

- `github.com/tmcrypted/berusmenu/...`

Local Go workspace:

- `go.work` joins `proto`, `api-gateway`, `auth-service`, and `notification-service` for local development.

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
- Web auth использует HttpOnly cookies для access/refresh tokens; frontend не хранит токены в JavaScript.
- Gateway выполняет cookie/CSRF проверки на HTTP-границе и передает identity во внутренние gRPC metadata.
- Внутренние вызовы между backend-сервисами идут по gRPC.
- Gateway stays thin; cross-domain business checks live in owning services.
- Kafka в MVP используется только для notification events.
- Redis в MVP используется для verification/rate limiting/presence.
- `notification-service` отделен от доменных сервисов и не имеет своей БД в MVP.
- Email channel в MVP: log/stub sender для local development; Mailganer planned as production provider.
- PVZ update directly mutates `inventory.pvz`, sets it back to `pending`, and is forbidden while active vacancy exists.
- Chat создается сразу после `response`.
- Chat становится read-only после `assignment.completed` и `assignment.cancelled`.

## Notification Flow

Kafka topics in MVP:

- `auth.email_verification_requested`
- `matching.response_created`

Notification delivery channel:

- email only

Delivery выполняет `notification-service`, который читает события из Kafka и передает их в email sender abstraction. В local development sender пишет email-заглушки в logs; Mailganer остается planned production provider.

Kafka bootstrap addresses:

- Docker compose services use `kafka:9092`.
- Host-side `go run` services use `localhost:19092`.

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

При регистрации gateway ставит auth cookies, но пока email не подтвержден, backend разрешает только whitelist endpoints:

- `logout`
- `resend verification code`
- `verify email`
- `me` is not available until email verification succeeds

Все остальные действия запрещены на backend, а frontend перенаправляет пользователя на экран подтверждения email.

## Web Auth Security

- `berusmenu_access` and `berusmenu_refresh` are HttpOnly cookies.
- `berusmenu_csrf` is readable by frontend and must be sent as `X-CSRF-Token` for authenticated unsafe requests.
- `Authorization: Bearer` is not used by the public web API.
- Refresh token revocation is stored in `auth.refresh_sessions` through hashed tokens and `revoked_at`.

## Documentation Map

- [AGENTS.md](AGENTS.md)
- [HTTP API](api-gateway/docs/http-api.md)
- [Insomnia Collection](api-gateway/insomnia/berusmenu-api.json)
- [Architecture](docs/architecture.md)
- [System Design](docs/system-design.md)
- [Sequence Diagrams](docs/sequence-diagrams.md)
- [State Transitions](docs/state-transitions.md)
- [Permissions](docs/permissions.md)
- [Interfaces](docs/interfaces.md)
- [Events And Cache](docs/events-and-cache.md)
- [Runtime Topology](docs/runtime-topology.md)
- [Development Plan](docs/development-plan.md)

## Что дальше

Следующие инженерные шаги:

1. Довести REST/gRPC contracts для `inventory-service`, `matching-service`, `communication-service`.
2. Реализовать `inventory-service`.
3. Реализовать `matching-service`.
4. Реализовать `communication-service`.
5. Доработать `api-gateway` REST -> gRPC mapping для доменных сервисов.
6. Реализовать frontend flows.
7. Добавить integration tests для основных business paths.
8. Описать production deployment notes в `infrastructure`.
