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

## Next implementation steps

1. PVZ CRUD
2. moderation endpoints
3. checks for vacancy creation and PVZ update restrictions
