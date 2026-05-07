# notification-service

Background worker for Berusmenu notification delivery.

## Ответственность

- consume Kafka notification events
- dispatch email notifications
- provide an email sender abstraction
- log delivery success and failure

## Storage

In MVP this service has no own PostgreSQL schema and no own DB.

## Consumed events

- `auth.email_verification_requested`
- `matching.response_created`

## Kafka processing policy

- Commit offsets only after successful handling.
- Malformed protobuf and semantically invalid notification events are logged and committed to avoid blocking a partition forever.
- Transient sender failures are retried with bounded backoff; if retries are exhausted, the offset is not committed.

## Delivery channel

MVP channel:

- email only
- local development uses a log/stub sender

Planned production provider:

- Mailganer

Later extensions:

- Telegram
- Max

## Components

- Kafka consumer
- event handlers
- email sender
- configuration
- external logging

## Internal structure

- `internal/consumer` owns Kafka transport: fetch, protobuf unmarshal, commit, and service dispatch.
- `internal/service` owns notification handling and the email stub sender abstraction for verification and response-created emails.
- `internal/dto` contains service-level notification models.
- Service behavior is covered by table-driven unit tests with gomock mocks.

## Next implementation steps

1. Mailganer email sender
2. handler for `matching.response_created`
3. delivery logs
