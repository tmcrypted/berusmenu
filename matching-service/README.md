# matching-service

`matching-service` owns the owner/worker marketplace workflow and is implemented as a runnable gRPC backend service.

## Responsibility

- vacancy lifecycle;
- response lifecycle;
- assignment lifecycle;
- reviews;
- review replies;
- active vacancy truth for PVZ update checks;
- response chat writability checks.
- reliable notification outbox for new responses;
- user review rating aggregates;
- audited moderator dispute actions.

## Owned Schema

Schema: `matching`

Tables:

- `matching.vacancies`
- `matching.responses`
- `matching.assignments`
- `matching.reviews`
- `matching.review_replies`
- `matching.user_ratings`
- `matching.outbox_events`
- `matching.admin_events`

## Rules

- Vacancy can be created only for approved owner PVZ.
- PVZ approval is checked through `inventory-service`, not by reading `inventory.*`.
- Active vacancy for inventory update guard means `published` or `assigned`.
- Repeat response to the same vacancy after `withdrawn` is forbidden.
- Assignment is created when owner accepts a response.
- Accepting a response uses the stored vacancy snapshot; it does not re-check current PVZ approval.
- Old vacancy is not reused after `assignment.cancelled`.
- Reviews are allowed after `completed` and `cancelled`.
- Review replies are flat: only one reply level under root review.
- Review reply `target_user_id` must be an assignment participant.
- `matching.response_created` is published after response creation.
- Response notification events are persisted to `matching.outbox_events` in the same transaction as response creation, then published by an outbox worker.
- Vacancy shift windows use full timestamps: `shift_start_at` and `shift_end_at`.
- `slots_total` must be `1` for MVP.
- User rating aggregates live in `matching.user_ratings`, not `auth.users`.
- Root review create/update recalculates the target user's rating synchronously in the same transaction.

## Required Inventory Integration

`matching-service` must implement:

```text
HasActiveVacancyForPvz(pvz_id)
```

Expected behavior:

- return `true` if at least one vacancy for the PVZ has status `published` or `assigned`;
- return `false` for `completed`, `cancelled`, or no vacancies;
- inventory uses this check before owner PVZ update and fails closed if the RPC is unavailable.

## Required Communication Integration

`matching-service` must expose:

```text
CheckResponseChatWritable(response_id)
```

`communication-service` will use it before accepting message writes.

## Moderator Integration

Public moderator HTTP routes are shaped under `/v1/moderation/matching/...` so a future moderation service can own cross-domain dispute cases without rewriting public API.

For MVP, matching state is still changed only by `matching-service`.

Rules:

- moderator can read all matching entities;
- moderator can perform business-valid dispute actions;
- moderator state-changing actions require an explicit `reason`;
- hard delete is not allowed;
- moderator actions must be auditable through `matching.admin_events`;
- admin audit writes happen in the same transaction as the matching state mutation.

## Internal Structure

- `cmd/matching-service` - process entrypoint.
- `internal/server` - gRPC adapter, metadata parsing, proto mapping and error mapping.
- `internal/service` - business rules and permission checks.
- `internal/dto` - service-layer input DTOs.
- `internal/store` - SQL queries, transactions and row mapping.
- `internal/clients` - typed auth and inventory clients.
- `internal/events` - outbox event builder and Kafka worker.
- `internal/migrations` - service-owned SQL migrations.

## Local Run

Run from service root so migrations resolve:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\matching-service
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
$env:INVENTORY_GRPC_ADDR="localhost:50052"
$env:AUTH_GRPC_ADDR="localhost:50051"
$env:KAFKA_BROKERS="localhost:19092"
$env:GRPC_ADDR=":50053"
go run .\cmd\matching-service
```

## Tests

```powershell
go test ./...
```

## Public Read Surfaces

The gateway exposes:

- role-aware vacancy listing;
- root review listing;
- flat reply listing sorted by `created_at ASC`;
- user rating aggregate reads from `matching.user_ratings`.
