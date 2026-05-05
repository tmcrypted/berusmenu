# proto

Директория общих protobuf contracts для backend-сервисов Berusmenu в рамках монорепозитория.

## Ответственность

- gRPC service definitions
- message contracts
- shared enums
- schema-first event payloads для Kafka notification events

## Текущее покрытие

- `auth.v1`
- `inventory.v1`
- `matching.v1`
- `communication.v1`
- `events.v1`
- `common.v1`

## Важные правила

- contracts уже выровнены под модель `vacancy/response/assignment/conversation/review_reply`
- старая vocabulary `listing/application/deal` больше не используется
- gateway сериализует protobuf ответы в snake_case proto field names
- event payloads описываются здесь, но не становятся частью public runtime API

## Kafka payload contracts

- `berusmenu.events.v1.EmailVerificationRequested`
- `berusmenu.events.v1.ResponseCreated`

## Tooling

- source files: `proto/**`
- generated Go stubs: `gen/go/**`

```bash
buf generate
```
