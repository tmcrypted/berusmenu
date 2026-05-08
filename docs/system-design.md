# System Design Berusmenu MVP

## Документ

Этот документ фиксирует итоговый design baseline для MVP. Подробности вынесены в специализированные разделы:

- [Sequence Diagrams](sequence-diagrams.md)
- [State Transitions](state-transitions.md)
- [Permissions](permissions.md)
- [Interfaces](interfaces.md)
- [Events And Cache](events-and-cache.md)
- [Runtime Topology](runtime-topology.md)
- [Development Plan](development-plan.md)

## Product Goal

Дать владельцу ПВЗ возможность быстро закрыть смену и выбрать исполнителя на конкретную дату, а worker — быстро откликнуться, договориться в чате и получить assignment.

## Must-have scope

- регистрация и логин по email/password
- HttpOnly cookie based web auth with CSRF protection
- email verification code
- PVZ registration and moderation
- vacancy creation
- worker response
- chat inside response flow
- assignment selection
- assignment cancellation/completion
- reviews
- review replies
- email notifications:
  - verification code
  - new response notification to owner

## Core Components

- `frontend`
- `api-gateway`
- `auth-service`
- `inventory-service`
- `matching-service`
- `communication-service`
- `notification-service`
- `postgres`
- `redis`
- `kafka`
- email sender abstraction

## Persisted Data

### `auth`

- `users`
- `refresh_sessions`

### `inventory`

- `pvz`
- `pvz_moderation_events`

### `matching`

- `vacancies`
- `responses`
- `assignments`
- `reviews`
- `review_replies`

### `communication`

- `conversations`
- `messages`
- `message_reads`

## Core Rules

- one worker response per vacancy
- conversation is created immediately after response
- repeat response after `withdrawn` is forbidden
- old vacancy is not reused after `assignment.cancelled`
- reviews are allowed after `completed` or `cancelled`
- replies belong only to root reviews and have one nesting level only
- chat is read-only after assignment finalization
- PVZ updates directly mutate `inventory.pvz`, set PVZ back to `pending`, and require repeated moderation
- PVZ update is forbidden when there are active vacancy with statuses `published` and `assigned`
- new vacancy cannot be created while PVZ is not `approved`
- Gateway stays thin; cross-domain business checks are synchronous RPCs owned by domain services

## Auth Rules

- user gets an access JWT and refresh token after registration, but gateway stores both in HttpOnly cookies for web clients
- frontend stores user/session state only, not raw tokens
- email may still be unverified
- unverified users can access only whitelist endpoints
- authenticated unsafe HTTP methods require `X-CSRF-Token` matching the readable CSRF cookie
- verification code is stored in Redis
- resend returns the same code while it is alive
- resend cooldown 60 seconds
- 3 failed verify attempts produce 1 hour lockout
- refresh sessions are revoked in Postgres by setting `auth.refresh_sessions.revoked_at`; only token hashes are persisted
- refresh rotation is atomic: revoke old session and insert new session in one transaction

## Async Rules

Kafka is used only for:

- `auth.email_verification_requested`
- `matching.response_created`

Consumer:

- `notification-service`

Delivery channel in MVP:

- email only
- local development uses log/stub sender
- Mailganer is planned as production provider

## Non-goals

- payments
- telegram/max notifications now
- advanced analytics
- tracing stack
- recommendation engine

## Implementation Status

System design is frozen enough for current auth/notification foundation. Next implementation focus:

1. inventory, matching and communication service implementation
2. api-gateway REST -> gRPC mapping for domain services
3. frontend flows
4. integration tests

## Frontend Auth Addendum

- registration can start with only `email + password + role`
- email verification remains mandatory before `GET /v1/me` and other verified-only flows
- after verification, frontend uses `GET /v1/me` to decide whether profile completion is still required
- profile completion is role-specific for `worker` and `owner`
- the first frontend slice is responsive for mobile, tablet, and desktop and supports manual light/dark theme switching
