# AGENTS

Global rules for the `berusmenu` monorepo.

## Scope

This file applies to every service and workspace in the monorepo.

If a service directory has its own `AGENTS.md`, it must:

- link to `../AGENTS.md` first;
- add only local rules;
- not override global product or architecture decisions.

## Product Context

Berusmenu is a marketplace for urgent PVZ shift staffing.

When designing and implementing, optimize for:

- short path from vacancy to assignment;
- simple operations;
- clean data ownership;
- fast MVP delivery to production.

## Service Boundaries

### `auth-service`

Owns accounts, sessions, email verification, JWT, profile and timed user blocks.

### `inventory-service`

Owns PVZ records and moderation workflow.

### `matching-service`

Owns vacancies, responses, assignments, reviews and review replies.

### `communication-service`

Owns conversations, messages, read state and chat semantics.

### `notification-service`

Owns Kafka notification consumption and email delivery.

### `api-gateway`

Owns public HTTP API, auth middleware and REST to gRPC mapping.

### `proto`

Owns protobuf contracts and shared enums.

## Data Rules

- One Postgres database, separate schema per service.
- FK only inside a service schema.
- Between schemas, store plain ids and validate through owning services.
- No service writes another service schema.

## Sync / Async Rules

- Core business transactions are synchronous.
- Kafka is used only for notification events in MVP.
- Redis is used for verification, rate limiting, presence and websocket ephemeral state.

Kafka topics in MVP:

- `auth.email_verification_requested`
- `matching.response_created`

## Auth Rules

- User receives JWT at registration.
- While `email_verified_at` is null, backend allows only whitelist endpoints.
- Verification code logic lives only in `auth-service`.
- Verification code is stored in Redis.
- Timed user block is represented by `auth.users.blocked_until`.

Redis keys:

- `email_verify:{user_id}`
- `email_verify_resend:{user_id}`
- `email_verify_attempts:{user_id}`

Verification policy:

- resend sends the same active code while it is alive;
- resend cooldown is 60 seconds;
- after 3 failed verify attempts, lockout is 1 hour.

## Matching Rules

- `response` can be `pending`, `accepted`, `rejected`, `withdrawn`.
- Repeat response to the same vacancy after `withdrawn` is forbidden.
- `assignment` is created when owner selects a worker.
- After `assignment.cancelled`, the old vacancy is not reused.
- Reviews are allowed after `completed` and `cancelled`.
- Replies are allowed only to root review and without nesting deeper than one level.

## Inventory Rules

- Vacancy can be created only for approved PVZ.
- Owner can edit PVZ.
- PVZ update is forbidden when active vacancy exists.
- Active vacancy for that rule means `published` or `assigned`.
- After update, PVZ becomes `pending` again.
- New vacancies for a PVZ are forbidden while PVZ is not `approved`.
- Moderator spam cleanup temporarily blocks owner, deletes pending owner PVZ requests and sets remaining owner PVZ to `blocked`.

## Chat Rules

- Conversation is created immediately after `response`.
- Chat is stored in `communication-service`.
- Chat is read-only after `assignment.completed` and `assignment.cancelled`.
- Only conversation participants can write to the conversation.

## Documentation Expectations

- All documentation must be written in English.
- Prefer ASCII in documentation files to avoid terminal/codepage issues.
- When API, contracts, runtime, env, Docker, service boundaries or developer workflow change, update the relevant `.md` files.

When product decisions change, update:

- root `README.md`;
- `docs/architecture.md`;
- `docs/system-design.md`.

When a concrete service changes, update its local docs.

When HTTP API changes in `api-gateway`, update:

- `api-gateway/docs/http-api.md`;
- `api-gateway/insomnia/berusmenu-api.json`.

## Non-Goals For MVP

Do not spend time on:

- payments;
- advanced observability stack;
- recommendation engine;
- CQRS/event sourcing for its own sake;
- ClickHouse analytics pipeline;
- multi-channel notifications beyond email.
