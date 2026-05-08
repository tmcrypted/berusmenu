# inventory-service

Сервис ПВЗ и модерации ПВЗ.

## Ответственность

- create PVZ
- read owner PVZ list
- update PVZ data
- approve/reject/block PVZ

## Owned schema

Schema: `inventory`

Tables:

- `inventory.pvz`
- `inventory.pvz_moderation_events`

## Важные правила

- vacancy can be created only for approved PVZ
- owner may edit PVZ data
- PVZ update is forbidden if there are active vacancy with status `published` or `assigned`
- after update PVZ returns to `pending`

## Integrations

- PostgreSQL
- gRPC calls from gateway and matching-service
- gRPC call to `matching-service.HasActiveVacancyForPvz` before PVZ update

## Responsibility split

- `inventory-service` owns `inventory.pvz`, `inventory.pvz_moderation_events`, PVZ permissions, PVZ state transitions, and moderation history.
- `matching-service` owns vacancy truth. Inventory never reads or writes `matching.*` directly.
- `api-gateway` owns only HTTP transport, auth/CSRF middleware, and REST -> gRPC mapping.

## Temporary matching dependency

`UpdatePvz` must reject changes while a PVZ has active vacancies with status `published` or `assigned`.

Until `matching-service` is implemented, inventory uses an `ActiveVacancyChecker` interface consumed by `internal/service`:

- runtime implementation calls `matching-service.HasActiveVacancyForPvz`
- tests use fakes/mocks
- failures are fail-closed: if matching is unavailable or returns an error, PVZ data is not mutated

When `matching-service` is implemented, add an end-to-end test that proves PVZ update is denied for active `published` and `assigned` vacancies.

## Implemented backend scope

1. PVZ create/read/list/update
2. moderation queue and approve/reject/block decisions
3. approved PVZ check for vacancy creation
4. active vacancy guard for PVZ updates through the matching interface
