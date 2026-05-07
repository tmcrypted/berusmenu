# matching-service

Сервис core marketplace flow для Berusmenu.

## Ответственность

- vacancy lifecycle
- response lifecycle
- assignment lifecycle
- reviews
- review replies

## Owned schema

Schema: `matching`

Tables:

- `matching.vacancies`
- `matching.responses`
- `matching.assignments`
- `matching.reviews`
- `matching.review_replies`

## Важные правила

- repeat response after `withdrawn` is forbidden
- vacancy creation validates approved owner PVZ through `inventory-service`
- active vacancy checks for PVZ updates are exposed to `inventory-service`
- assignment is created from accepted response
- response creation synchronously creates conversation through `communication-service`
- old vacancy is not reused after `assignment.cancelled`
- reviews are allowed after `completed` and `cancelled`
- review replies are flat and sorted by `created_at ASC`

## Integrations

- PostgreSQL
- gRPC via gateway
- gRPC calls to `inventory-service` for approved PVZ validation
- calls to `communication-service` for conversation creation
- gRPC checks for `inventory-service` and `communication-service`
- Kafka producer for `matching.response_created`

## Next implementation steps

1. vacancy CRUD
2. response create/withdraw/accept
3. assignment create/cancel/complete
4. reviews and replies
