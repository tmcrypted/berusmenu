# Kafka Events and Redis Usage

## Kafka in MVP

Kafka is used only for notification events.

### Event: `auth.email_verification_requested`

Producer:

- `auth-service`

Consumer:

- `notification-service`

Payload:

- `user_id`
- `email`
- `code`
- `occurred_at`

Side effect:

- send verification email through Mailganer

### Event: `matching.response_created`

Producer:

- `matching-service`

Consumer:

- `notification-service`

Payload:

- `response_id`
- `vacancy_id`
- `owner_user_id`
- `owner_email`
- `worker_user_id`
- `occurred_at`

Side effect:

- send owner email about new response

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

### Presence / realtime state

- online/offline presence
- typing status
- active websocket connections
