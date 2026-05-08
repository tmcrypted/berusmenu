# notification-service

`notification-service` is a background worker for notification delivery.

It has no public HTTP/gRPC API and no PostgreSQL schema in MVP.

## Responsibility

- consume Kafka notification events;
- decode protobuf event payloads;
- dispatch email notifications;
- provide email sender abstraction;
- log delivery success/failure.

## Consumed Events

- `auth.email_verification_requested`
- `matching.response_created`

## Kafka Processing Policy

- Commit offsets only after successful handling.
- Malformed protobuf and semantically invalid messages are logged and committed so a partition is not blocked forever.
- Transient sender failures are retried with bounded backoff.
- If retries are exhausted, the offset is not committed.

## Delivery Channel

MVP:

- email only;
- local development uses log/stub sender.

Planned:

- production email provider;
- optional Telegram/Max later.

## Internal Structure

- `internal/consumer` - Kafka fetch/unmarshal/commit transport.
- `internal/service` - notification handling and email sender abstraction.
- `internal/dto` - service-level notification models.
- `internal/config` - env parsing.

## Local Run

From service root:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\notification-service
$env:KAFKA_BROKERS="localhost:19092"
go run .\cmd\notification-service
```

Inside compose use:

```text
KAFKA_BROKERS=kafka:9092
```

## Tests

```powershell
go test ./...
```

## Next Steps

1. Production email sender.
2. Delivery logging/metrics.
3. Real `matching.response_created` producer integration after matching-service is implemented.
