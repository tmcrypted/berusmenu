# Kafka Events and Redis Usage

## Kafka in MVP

Kafka is used only for notification events.

`auth-service` writes verification notification events to `auth.outbox_events` before Kafka publication. Its outbox worker publishes pending rows asynchronously, so Kafka downtime does not lose verification events.

`notification-service` commits Kafka offsets only after successful handling. Malformed protobuf, missing `event_id`, or semantically invalid notification messages are logged and committed so they do not block a partition forever. Transient sender failures are retried with bounded backoff and are not committed after retries are exhausted. Duplicate `event_id` values are committed without sending duplicate email.

### Event: `auth.email_verification_requested`

Producer:

- `auth-service`

Consumer:

- `notification-service`

Payload:

- `user_id`
- `event_id`
- `email`
- `code`
- `occurred_at`

Side effect:

- send verification email through notification email sender
- local development currently logs an email stub
- Mailganer is planned as production provider

### Event: `matching.response_created`

Producer:

- `matching-service`

Producer behavior:

- response creation and `matching.response_created` outbox insert happen in one database transaction;
- an outbox worker publishes pending matching events to Kafka.

Consumer:

- `notification-service`

Payload:

- `response_id`
- `event_id`
- `vacancy_id`
- `owner_user_id`
- `owner_email`
- `worker_user_id`
- `vacancy_title`
- `shift_date`
- `pay_amount`
- `occurred_at`

Side effect:

- send owner email about new response through notification email sender

Consumer behavior:

- both MVP topics are consumed by `notification-service` in the same consumer group
- local development uses the log/stub sender for both notification types

## Redis in MVP

### Verification code

- `email_verify:{user_id}` -> active verification code, TTL 10m

### Resend cooldown

- `email_verify_resend:{user_id}` -> resend cooldown, TTL 60s

### Verify attempts

- `email_verify_attempts:{user_id}` -> failed attempts counter / lockout logic

Policy:

- same active code is resent while alive
- after 3 failed attempts -> lockout 1 hour

### Rate limiting

Expected categories:

- register
- login
- resend verification code
- verify email
- send message

Current implemented Redis usage is email verification. Rate limiting and realtime state are reserved for the following service slices.

### Presence / realtime state

- online/offline presence
- typing status
- active websocket connections
