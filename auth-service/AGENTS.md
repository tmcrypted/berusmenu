# AGENTS

See the global guide first: [../AGENTS.md](../AGENTS.md)

## Service Role

`auth-service` owns identity, sessions and email verification.

## Local Rules

- Verification code generation and verification live here.
- Redis is an implementation detail of auth verification and rate-limit logic.
- Only auth-service updates `email_verified_at`.
- JWT must contain enough identity data for gateway restrictions.
