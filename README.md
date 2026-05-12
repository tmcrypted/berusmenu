# Berusmenu MVP

Berusmenu is a marketplace for urgent PVZ shift staffing. The MVP goal is to give a PVZ owner a short path from an uncovered shift to a selected worker, while keeping operations and service ownership simple.

This monorepo contains backend services, frontend, protobuf contracts and local infrastructure.

Canonical Go import path:

```text
github.com/tmcrypted/berusmenu/...
```

Local Go workspace:

```text
go.work
```

Current `go.work` modules:

- `proto`
- `api-gateway`
- `auth-service`
- `inventory-service`
- `matching-service`
- `notification-service`

## Repository Structure

- `api-gateway` - public HTTP API, cookie auth, CSRF, REST to gRPC mapping.
- `auth-service` - accounts, sessions, JWT, email verification, profile, timed user block.
- `inventory-service` - PVZ records, moderation statuses, moderation history, owner checks.
- `matching-service` - vacancy, response, assignment, review, rating and matching moderation workflow.
- `communication-service` - planned service for conversations/messages/read state.
- `notification-service` - Kafka consumer and email notification delivery.
- `proto` - protobuf contracts and shared enums.
- `frontend` - Next.js frontend.
- `infrastructure` - Docker Compose local runtime.
- `docs` - architecture, permissions, state transitions, runtime topology and roadmap.

## Current Status

Implemented:

- `auth-service`: register/login/refresh/logout, JWT validation, `GET/PATCH /v1/me`, email verification, resend cooldown/lockout, timed user block through `blocked_until`.
- `notification-service`: consumes `auth.email_verification_requested` and `matching.response_created`, deduplicates by `event_id`, uses retry-aware Kafka commit policy and a local email stub sender.
- `inventory-service`: create/read/list/update PVZ, moderation queue, approve/reject/block PVZ, owner-only access, moderator-only moderation, temporary matching dependency for active vacancy checks.
- `matching-service`: vacancy create/list/get, response create/withdraw/list/accept, assignment cancel/complete, reviews/replies, rating aggregates, active vacancy and chat writability checks, response notification outbox.
- `api-gateway`: auth/profile routes, inventory routes, moderation routes, cookie/CSRF middleware, gRPC clients.
- `frontend`: auth flow, email verification, profile completion with optional messenger usernames, guest vacancy feed, owner PVZ flow, moderator PVZ moderation panel, timed owner spam block action, SSR matching vacancy/response/assignment/review screens, and read-only chat placeholders.

Not implemented as runnable services yet:

- `communication-service`

They already have protobuf outlines, architecture rules and Docker future-profile placeholders.

## Product Flow

1. User registers with email/password and role.
2. User verifies email by code.
3. User completes profile after verification.
4. Owner creates PVZ; the request enters `pending`.
5. Moderator approves, rejects or blocks PVZ.
6. Owner can create vacancy only for approved PVZ.
7. Worker responds to vacancy.
8. Conversation is created after response.
9. Owner selects worker and assignment is created.
10. After `completed` or `cancelled`, both sides can create review and reply.

The current implemented slice covers backend matching flows and frontend screens for vacancies, responses, assignments, reviews and read-only chat placeholders. The real communication service and realtime chat UI are next.

## Architecture Rules

- One PostgreSQL database, separate schema per service.
- FK only inside a service schema.
- Cross-schema references are stored as plain ids and validated through owning services.
- No service writes another service schema.
- Public web API goes through `api-gateway`.
- Browser auth uses HttpOnly cookies for access/refresh tokens.
- Frontend does not store JWT or refresh token in JavaScript.
- Authenticated unsafe HTTP methods require `X-CSRF-Token`.
- Gateway forwards identity to internal gRPC calls through metadata.
- Gateway stays thin; business rules live in owning services.
- Kafka is used only for notification events in MVP.
- Notification events are published through service-owned outbox tables.
- Redis is used for verification, rate limiting and future realtime ephemeral state.

## Data Ownership

### `auth-service`

Schema: `auth`

- `auth.users`
- `auth.refresh_sessions`
- `auth.outbox_events`

`auth.users.blocked_until` supports temporary moderator blocks. While a block is active, login/refresh/JWT validation are denied. After `blocked_until` expires, the user can authenticate again without a manual unblock.

### `inventory-service`

Schema: `inventory`

- `inventory.pvz`
- `inventory.pvz_moderation_events`

PVZ statuses:

- `pending`
- `approved`
- `rejected`
- `blocked`

Owner can update their PVZ, but update returns PVZ to `pending` and is denied when active vacancy exists. Active vacancy means future matching status `published` or `assigned`.

Moderator spam cleanup:

- temporarily blocks owner in `auth-service`;
- deletes owner pending PVZ requests;
- sets owner remaining PVZ to `blocked`;
- writes `blocked` moderation events for affected PVZ rows.

### `matching-service`

Planned schema: `matching`

- `matching.vacancies`
- `matching.responses`
- `matching.assignments`
- `matching.reviews`
- `matching.review_replies`

This service will own vacancy truth and `HasActiveVacancyForPvz`.

### `communication-service`

Planned schema: `communication`

- `communication.conversations`
- `communication.messages`
- `communication.message_reads`

This service will own chat semantics.

### `notification-service`

Schema: `notification`

- `notification.processed_events`

It runs as a background worker and stores processed event ids for idempotency.

## Runtime

Default Docker Compose stack:

- `postgres`
- `redis`
- `kafka` / Redpanda
- `auth-service`
- `inventory-service`
- `notification-service`
- `api-gateway`
- `frontend`

Future profile:

- `matching-service`
- `communication-service`

Inside Compose, services use:

```text
postgres:5432
redis:6379
kafka:9092
```

For host-side `go run`, use host ports:

```powershell
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:KAFKA_BROKERS="localhost:19092"
```

Run Go services from the service root so `MIGRATIONS_DIR=internal/migrations` resolves correctly.

Auth:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\auth-service
$env:GRPC_ADDR=":50051"
go run .\cmd\auth-service
```

Inventory:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\inventory-service
$env:GRPC_ADDR=":50052"
$env:MATCHING_GRPC_ADDR="localhost:50053"
go run .\cmd\inventory-service
```

Gateway:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\api-gateway
$env:HTTP_ADDR=":8080"
$env:AUTH_GRPC_ADDR="localhost:50051"
$env:INVENTORY_GRPC_ADDR="localhost:50052"
go run .\cmd\api-gateway
```

Notification:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\notification-service
$env:KAFKA_BROKERS="localhost:19092"
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
go run .\cmd\notification-service
```

## Public HTTP Scope

Auth/profile:

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `GET /v1/auth/session`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

Inventory:

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/{id}`
- `PATCH /v1/pvz/{id}`
- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/{id}/approve`
- `POST /v1/moderation/pvz/{id}/reject`
- `POST /v1/moderation/pvz/{id}/block`
- `POST /v1/moderation/users/{id}/block-and-delete-pvz`

Matching and communication routes are placeholders until those services are implemented.

## Kafka Events

MVP topics:

- `auth.email_verification_requested`
- `matching.response_created`

`auth-service` first persists verification events in `auth.outbox_events`, then an internal worker publishes them to Kafka. `notification-service` consumes both topics and deduplicates by `event_id`. Local delivery uses a log/stub sender.

## Redis Keys

- `email_verify:{user_id}` - verification code, TTL 10 minutes.
- `email_verify_resend:{user_id}` - resend cooldown, TTL 60 seconds.
- `email_verify_attempts:{user_id}` - failed attempts and 1 hour lockout after 3 failures.

Planned Redis usage also includes rate limiting, presence, typing state and websocket ephemeral state.

## Verification

Backend:

```powershell
go test ./api-gateway/... ./auth-service/... ./inventory-service/... ./notification-service/... ./proto/...
```

Frontend:

```powershell
cd frontend
npm run lint
npx tsc --noEmit
```

Production build is not required for every local check.

## Documentation Map

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
- [Infrastructure](infrastructure/README.md)
- [Frontend](frontend/README.md)

When HTTP API behavior changes, update both:

- `api-gateway/docs/http-api.md`
- `api-gateway/insomnia/berusmenu-api.json`

## Next Steps

1. Implement `matching-service` backend.
2. Replace the temporary inventory active-vacancy dependency with real matching integration.
3. Implement vacancy/response/assignment frontend and backend flows.
4. Implement `communication-service` and chat UI.
5. Add integration tests for main business paths.
6. Document production deployment in `infrastructure`.
