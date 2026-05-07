# auth-service

Сервис аутентификации и пользовательского профиля для Berusmenu.

## Ответственность

- register
- login
- refresh/logout
- JWT validation
- current user profile
- email verification
- resend verification code

## Owned schema

Schema: `auth`

Tables:

- `auth.users`
- `auth.refresh_sessions`

## Важные правила

- `email_verified_at` controls verified access
- user gets JWT at registration time
- raw refresh tokens are never stored; only `SHA-256` token hashes are persisted
- refresh revocation uses `auth.refresh_sessions.revoked_at`
- refresh rotation is atomic in Postgres: validate old active session, revoke it, and insert the new session in one transaction
- verification code logic lives here
- Redis keys for verification are owned logically by this service
- avatars are stored on server files for now, DB stores `avatar_url`

## Integrations

- PostgreSQL
- Redis
- Kafka producer for `auth.email_verification_requested`

## Internal structure

- `internal/server` is a thin gRPC adapter: protobuf mapping, metadata parsing, and service error mapping.
- `internal/service` owns auth business logic and is covered by table-driven unit tests with gomock mocks.
- `internal/dto` contains service-level input/output models that are independent from protobuf contracts.
- `internal/store`, `internal/verification`, and `internal/events` remain infrastructure adapters for PostgreSQL, Redis, and Kafka.
- `internal/store` uses Squirrel for runtime SQL construction; migrations remain raw SQL.
- gRPC health service is registered for runtime readiness/liveness integration.

## Next implementation steps

1. user model and repository
2. register/login/refresh/logout
3. verify email flow
4. resend cooldown and attempts lockout
5. JWT validation for gateway
