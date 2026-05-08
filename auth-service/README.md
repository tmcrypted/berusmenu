# auth-service

`auth-service` owns accounts, sessions, JWT validation, email verification and user profile data.

## Responsibility

- register;
- login;
- refresh/logout;
- JWT validation for gateway;
- current user profile;
- email verification;
- resend verification code;
- temporary moderator block for spam users.

## Owned Schema

Schema: `auth`

Tables:

- `auth.users`
- `auth.refresh_sessions`

Important `auth.users` fields:

- `email`
- `email_verified_at`
- `password_hash`
- `role`
- `full_name`
- `phone`
- `avatar_url`
- `status`
- `blocked_until`

`blocked_until` is nullable. While `status = blocked` and `blocked_until` is in the future, login/refresh/JWT validation are denied. After `blocked_until` expires, the user can authenticate again without a manual unblock.

## Rules

- User gets JWT and refresh token at registration time.
- Until `email_verified_at` is set, gateway allows only verification/auth whitelist routes.
- Raw refresh tokens are never stored; only SHA-256 hashes are persisted.
- Refresh rotation is atomic: validate old active session, revoke it, insert new session.
- Logout revokes refresh session through `auth.refresh_sessions.revoked_at`.
- Verification code logic lives only in this service.
- Verification code state is stored in Redis.
- Moderator cannot block themself or another moderator.

## Redis Keys

- `email_verify:{user_id}` - code, TTL 10 minutes.
- `email_verify_resend:{user_id}` - resend cooldown, TTL 60 seconds.
- `email_verify_attempts:{user_id}` - failed attempts and 1 hour lockout after 3 failures.

## Kafka

Producer:

- `auth.email_verification_requested`

Payload is defined in `proto/events/v1/notifications.proto`.

## gRPC

Implemented:

- `Register`
- `Login`
- `Refresh`
- `Logout`
- `ValidateToken`
- `GetUser`
- `UpdateProfile`
- `ResendVerificationCode`
- `VerifyEmail`
- `BlockUser`

`BlockUser` consumes moderator identity metadata from gateway and requires `blocked_until`.

## Internal Structure

- `internal/server` - thin gRPC adapter.
- `internal/service` - business rules and testable logic.
- `internal/dto` - service-level DTOs.
- `internal/store` - PostgreSQL adapter.
- `internal/verification` - Redis verification adapter.
- `internal/events` - Kafka producer.
- `internal/migrations` - service-owned SQL migrations.

## Local Run

Run from service root so migrations resolve:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\auth-service
$env:POSTGRES_DSN="postgres://berusmenu:berusmenu@localhost:54320/berusmenu?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:KAFKA_BROKERS="localhost:19092"
$env:GRPC_ADDR=":50051"
go run .\cmd\auth-service
```

Inside Docker compose the service uses `postgres:5432`, `redis:6379`, `kafka:9092`.

## Tests

```powershell
go test ./...
```
