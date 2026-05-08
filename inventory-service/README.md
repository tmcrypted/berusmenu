# inventory-service

`inventory-service` owns PVZ data and PVZ moderation workflow.

## Responsibility

- create PVZ;
- read one PVZ;
- list owner PVZ;
- update owner PVZ data;
- check approved PVZ for future vacancy creation;
- list pending PVZ for moderators;
- approve/reject/block PVZ;
- moderator owner spam cleanup.

## Owned Schema

Schema: `inventory`

Tables:

- `inventory.pvz`
- `inventory.pvz_moderation_events`

No FK points to other service schemas. `owner_id`, `moderated_by`, and `actor_user_id` are id references to auth users, validated through service boundaries and auth metadata.

## PVZ Statuses

- `pending`
- `approved`
- `rejected`
- `blocked`

Transitions:

- create -> `pending`;
- approve -> `approved`;
- reject -> `rejected`;
- block -> `blocked`;
- owner update -> `pending`;
- spam cleanup -> pending rows deleted, non-pending rows set to `blocked`.

## Rules

- Owner can create/list/read/update only their own PVZ.
- Moderator can inspect moderation queue and moderate PVZ.
- `RejectPvz` requires reason.
- Every moderation decision appends a moderation event.
- Owner update resets moderation decision fields and appends `submitted`.
- Owner update is forbidden while active vacancy exists.
- Vacancy can be created only for approved PVZ.
- Spam cleanup deletes only pending PVZ requests for the owner and sets remaining owner PVZ to `blocked`.

## Temporary Matching Dependency

`UpdatePvz` calls `matching-service.HasActiveVacancyForPvz` through the `ActiveVacancyChecker` interface.

Rules:

- active vacancy means status `published` or `assigned`;
- if matching reports active vacancy, update is denied;
- if matching is unavailable, update fails closed and does not mutate PVZ.

Until `matching-service` is implemented, runtime still points to `MATCHING_GRPC_ADDR`, and tests use fakes/mocks. After matching implementation, add end-to-end coverage for the active-vacancy guard.

## gRPC

Implemented:

- `CreatePvz`
- `GetPvz`
- `ListOwnerPvz`
- `UpdatePvz`
- `CheckPvzForVacancy`
- `ListPendingPvz`
- `ApprovePvz`
- `RejectPvz`
- `BlockPvz`
- `DeleteOwnerPvzRequests`

`DeleteOwnerPvzRequests` currently implements spam cleanup semantics: delete pending owner PVZ requests and block remaining owner PVZ.

## Internal Structure

- `internal/server` - thin gRPC adapter and metadata parsing.
- `internal/service` - business rules.
- `internal/dto` - service-level DTOs.
- `internal/store` - SQL queries, transactions and scan/map logic.
- `internal/matching` - active vacancy checker adapter.
- `internal/migrations` - service-owned SQL migrations.

## Local Run

Run from service root so migrations resolve:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\inventory-service
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
$env:MATCHING_GRPC_ADDR="localhost:50053"
$env:GRPC_ADDR=":50052"
go run .\cmd\inventory-service
```

## Tests

```powershell
go test ./...
```
