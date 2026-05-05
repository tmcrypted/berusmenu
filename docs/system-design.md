# System Design Berusmenu MVP

## Документ

Этот документ фиксирует итоговый design baseline для MVP. Подробности вынесены в специализированные разделы:

- [Sequence Diagrams](C:/Users/tmcrypted/Desktop/berusmenu/docs/sequence-diagrams.md)
- [State Transitions](C:/Users/tmcrypted/Desktop/berusmenu/docs/state-transitions.md)
- [Permissions](C:/Users/tmcrypted/Desktop/berusmenu/docs/permissions.md)
- [Interfaces](C:/Users/tmcrypted/Desktop/berusmenu/docs/interfaces.md)
- [Events And Cache](C:/Users/tmcrypted/Desktop/berusmenu/docs/events-and-cache.md)
- [Runtime Topology](C:/Users/tmcrypted/Desktop/berusmenu/docs/runtime-topology.md)
- [Development Plan](C:/Users/tmcrypted/Desktop/berusmenu/docs/development-plan.md)

## Product Goal

Дать владельцу ПВЗ возможность быстро закрыть смену и выбрать исполнителя на конкретную дату, а worker — быстро откликнуться, договориться в чате и получить assignment.

## Must-have scope

- регистрация и логин по email/password
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
- `Mailganer`

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
- PVZ updates require repeated moderation
- submit PVZ update is forbidden when there are active vacancy with statuses `published` and `assigned`
- new vacancy cannot be created while PVZ is not `approved`

## Auth Rules

- user gets JWT after registration
- email may still be unverified
- unverified users can access only whitelist endpoints
- verification code is stored in Redis
- resend returns the same code while it is alive
- resend cooldown 60 seconds
- 3 failed verify attempts produce 1 hour lockout

## Async Rules

Kafka is used only for:

- `auth.email_verification_requested`
- `matching.response_created`

Consumer:

- `notification-service`

Delivery channel in MVP:

- email only via Mailganer

## Non-goals

- payments
- telegram/max notifications now
- advanced analytics
- tracing stack
- recommendation engine

## Implementation Status

System design is frozen enough to move into:

1. protobuf contracts
2. goose migrations
3. service implementation
