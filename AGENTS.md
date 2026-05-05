# AGENTS

Глобальный файл правил для монорепозитория `berusmenu`.

## Область действия

Этот файл обязателен для всех сервисов внутри монорепозитория.

Если в директории сервиса есть свой `AGENTS.md`, он:

- сначала ссылается на `../AGENTS.md`
- дополняет только локальные правила
- не переопределяет глобальные продуктовые и архитектурные решения

## Продуктовый контекст

Berusmenu — маркетплейс срочного подбора сотрудника на смену в ПВЗ.

При проектировании и реализации оптимизируем за:

- короткий путь от vacancy к assignment
- простую эксплуатацию
- чистое владение данными
- быстрый вывод MVP в production

## Сервисные границы

### `auth-service`

Владеет account, sessions, email verification, JWT и profile.

### `inventory-service`

Владеет PVZ и moderation workflow.

### `matching-service`

Владеет vacancies, responses, assignments, reviews и review replies.

### `communication-service`

Владеет conversations, messages, read state и chat semantics.

### `notification-service`

Владеет Kafka notification consumption и email delivery.

### `api-gateway`

Владеет public HTTP API, auth middleware и REST to gRPC mapping.

### `proto`

Владеет protobuf contracts и enums.

## Data Rules

- Одна база Postgres, отдельная схема на сервис.
- FK только внутри схем сервиса.
- Между схемами только id references и валидация в owning service.
- Один сервис не пишет в схему другого сервиса.

## Sync / Async Rules

- Основные бизнес-транзакции синхронные.
- Kafka в MVP используется только для notification events.
- Redis в MVP используется только для verification/rate limiting/presence and websocket ephemeral state.

Kafka topics in MVP:

- `auth.email_verification_requested`
- `matching.response_created`

## Auth Rules

- Пользователь получает JWT при регистрации.
- Пока `email_verified_at` is null, backend разрешает только whitelist endpoints.
- Verification code logic живет только в `auth-service`.
- Verification code хранится в Redis.

Redis keys:

- `email_verify:{user_id}`
- `email_verify_resend:{user_id}`
- `email_verify_attempts:{user_id}`

Verification policy:

- resend шлет тот же код, пока он еще жив
- resend cooldown 60 seconds
- после 3 failed verify attempts lockout 1 hour

## Matching Rules

- `response` can be `pending`, `accepted`, `rejected`, `withdrawn`
- повторный отклик на ту же vacancy после `withdrawn` запрещен
- `assignment` создается при выборе worker
- после `assignment.cancelled` старая vacancy больше не используется
- reviews разрешены после `completed` и `cancelled`
- replies разрешены только к root review и без вложенности глубже одного уровня

## Inventory Rules

- vacancy можно создать только для approved PVZ
- owner может редактировать ПВЗ
- submit update request for PVZ запрещен при active vacancy
- active vacancy для этого правила: `published` и `assigned`
- после submit update ПВЗ снова становится `pending`
- новые vacancy по ПВЗ запрещены, пока ПВЗ не `approved`

## Chat Rules

- conversation создается сразу после `response`
- chat хранится в `communication-service`
- chat read-only после `assignment.completed` и `assignment.cancelled`
- писать в conversation могут только его участники

## Documentation Expectations

При изменении продуктовых решений обновлять:

- root `README.md`
- `docs/architecture.md`
- `docs/system-design.md`

При изменении конкретного сервиса обновлять его локальные docs.

## Non-goals for MVP

Пока не тратим время на:

- payments
- advanced observability stack
- recommendation engine
- CQRS/event sourcing ради самих себя
- clickhouse analytics pipeline
- multi-channel notifications beyond email
