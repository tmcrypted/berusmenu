# proto

Shared protobuf contracts for Berusmenu backend services.

## Responsibility

- gRPC service definitions;
- message contracts;
- shared enums;
- schema-first Kafka event payloads.

## Packages

- `auth.v1`
- `inventory.v1`
- `matching.v1`
- `communication.v1`
- `events.v1`
- `common.v1`

## Current Notes

- `auth.v1.AuthService` includes `BlockUser` for timed moderator blocks.
- `inventory.v1.InventoryService` includes `DeleteOwnerPvzRequests` for moderator spam cleanup.
- Gateway serializes protobuf HTTP responses using proto field names in snake_case.
- Kafka event payloads are described here, but they are not public HTTP API.
- Old vocabulary such as `listing/application/deal` is not used; current domain language is `vacancy/response/assignment`.

## Kafka Payload Contracts

- `berusmenu.events.v1.EmailVerificationRequested`
- `berusmenu.events.v1.ResponseCreated`

All notification events carry `event_id` for at-least-once delivery deduplication.

## Tooling

Source files:

```text
proto/**
```

Generated Go stubs:

```text
gen/go/**
```

Regenerate:

```powershell
buf generate
```
