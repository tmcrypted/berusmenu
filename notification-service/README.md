# notification-service

Background worker for Berusmenu notification delivery.

## Ответственность

- consume Kafka notification events
- dispatch email notifications
- integrate with Mailganer
- log delivery success and failure

## Storage

In MVP this service has no own PostgreSQL schema and no own DB.

## Consumed events

- `auth.email_verification_requested`
- `matching.response_created`

## Delivery channel

MVP channel:

- email only

Later extensions:

- Telegram
- Max

## Components

- Kafka consumer
- event handlers
- email sender
- configuration
- external logging

## Next implementation steps

1. config and bootstrap
2. kafka consumer
3. Mailganer email sender
4. handlers for two events
5. delivery logs
