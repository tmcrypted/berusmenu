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
- verification code logic lives here
- Redis keys for verification are owned logically by this service
- avatars are stored on server files for now, DB stores `avatar_url`

## Integrations

- PostgreSQL
- Redis
- Kafka producer for `auth.email_verification_requested`

## Next implementation steps

1. user model and repository
2. register/login/refresh/logout
3. verify email flow
4. resend cooldown and attempts lockout
5. JWT validation for gateway
