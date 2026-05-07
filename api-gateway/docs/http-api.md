# Berusmenu HTTP API

This document is the maintained HTTP API reference for `api-gateway`.

Base URL for local Docker:

```text
http://localhost:8080
```

## Conventions

Web auth uses cookies set by the gateway:

```text
Cookie: berusmenu_access=<access_jwt>
Cookie: berusmenu_refresh=<refresh_token>
Cookie: berusmenu_csrf=<csrf_token>
```

`berusmenu_access` and `berusmenu_refresh` are `HttpOnly`; frontend JavaScript cannot read them. `Authorization: Bearer` is not accepted by the web API.

Authenticated unsafe methods (`POST`, `PATCH`, `PUT`, `DELETE`) require:

```text
X-CSRF-Token: <value from berusmenu_csrf cookie>
```

The maintained Insomnia collection reads this value from the cookie jar with:

```text
{% cookie _.base_url, 'berusmenu_csrf' %}
```

Most business endpoints also require verified email. The gateway allows unverified authenticated users only for:

- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`

Errors use this shape:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "validation failed",
    "request_id": "req-123"
  }
}
```

Common error codes:

- `400 invalid_argument`
- `401 unauthenticated`
- `403 permission_denied`
- `409 already_exists`
- `412 failed_precondition`
- `429 resource_exhausted`
- `500 internal`
- `501 not_implemented`
- `503 unavailable`
- `504 deadline_exceeded`

## Health

### `GET /health/live`

Process liveness.

Success `200`:

```json
{
  "status": "ok"
}
```

### `GET /health/ready`

Traffic readiness.

Success `200`:

```json
{
  "status": "ok"
}
```

## Auth

### `POST /v1/auth/register`

Creates a user account, creates a refresh session, publishes an email verification event, sets auth cookies, and returns user/session metadata.

Auth: none.

Request:

```json
{
  "email": "worker@example.com",
  "password": "password123",
  "full_name": "Worker User",
  "role": "worker",
  "phone": "+79990001122"
}
```

Fields:

- `email`: required, email.
- `password`: required, minimum 8 chars.
- `full_name`: required.
- `role`: optional, one of `worker`, `owner`, `moderator`; defaults to `worker`.
- `phone`: optional.

Success `201`:

```json
{
  "user": {
    "id": "7",
    "email": "worker@example.com",
    "role": "USER_ROLE_WORKER",
    "full_name": "Worker User",
    "created_at": "2026-05-06T13:00:00Z",
    "updated_at": "2026-05-06T13:00:00Z"
  },
  "access_expires_at": "2026-05-06T13:15:00Z"
}
```

Response sets `berusmenu_access`, `berusmenu_refresh`, and `berusmenu_csrf` cookies.

Possible errors: `400`, `409`, `503`.

### `POST /v1/auth/login`

Authenticates a user, sets auth cookies, and returns user/session metadata.

Auth: none.

Request:

```json
{
  "email": "worker@example.com",
  "password": "password123"
}
```

Success `200`: same shape as register success.

Possible errors: `400`, `401`, `403`.

### `POST /v1/auth/refresh`

Exchanges the refresh cookie for new auth cookies and user/session metadata.

Auth: none.

Request: no body required.

```json
{}
```

Success `200`: same shape as register success.

Possible errors: `400`, `401`, `403`.

### `POST /v1/auth/logout`

Revokes a refresh token.

Auth: refresh cookie required. CSRF header required. Allowed before email verification and after access expiry.

Request: no body required.

```json
{}
```

Success `204`: no body. Response clears auth and CSRF cookies.

Possible errors: `400`, `401`, `500`.

### `POST /v1/auth/resend-code`

Resends the current email verification code when cooldown allows.

Auth: access cookie required. CSRF header required. Allowed before email verification.

Request: empty JSON body is accepted.

```json
{}
```

Success `200`:

```json
{
  "code_expires_at": "2026-05-06T13:10:00Z",
  "resend_available_at": "2026-05-06T13:01:00Z"
}
```

Possible errors: `401`, `412`, `429`, `503`.

### `POST /v1/auth/verify-email`

Verifies the current user's email.

Auth: access cookie required. CSRF header required. Allowed before email verification.

Request:

```json
{
  "code": "123456"
}
```

Success `200`:

```json
{
  "user": {
    "id": "7",
    "email": "worker@example.com",
    "email_verified_at": "2026-05-06T13:02:00Z",
    "role": "USER_ROLE_WORKER",
    "full_name": "Worker User",
    "created_at": "2026-05-06T13:00:00Z",
    "updated_at": "2026-05-06T13:02:00Z"
  }
}
```

Possible errors: `400`, `401`, `429`.

Local development note: the verification code is printed by `auth-service` in local/dev mode and by `notification-service` when Kafka delivery is running. Copy that code into the Insomnia `verification_code` environment variable before running this request.

## Profile

### `GET /v1/me`

Returns the authenticated user's profile.

Auth: access cookie required. Requires verified email.

Success `200`:

```json
{
  "id": "7",
  "email": "worker@example.com",
  "email_verified_at": "2026-05-06T13:02:00Z",
  "role": "USER_ROLE_WORKER",
  "full_name": "Worker User",
  "created_at": "2026-05-06T13:00:00Z",
  "updated_at": "2026-05-06T13:02:00Z"
}
```

Possible errors: `401`, `403`, `404`.

### `PATCH /v1/me`

Updates provided profile fields.

Auth: access cookie required. CSRF header required. Requires verified email.

Request:

```json
{
  "full_name": "New Name",
  "phone": "+79990001122",
  "avatar_url": "https://example.com/avatar.png"
}
```

All fields are optional. Omitted fields are unchanged.

Success `200`: same shape as `GET /v1/me`.

Possible errors: `400`, `401`, `403`, `404`.

## Placeholder Endpoints

These routes are registered in the gateway but currently return:

```json
{
  "error": {
    "code": "not_implemented",
    "message": "... gateway routes are not implemented yet",
    "request_id": "req-123"
  }
}
```

Inventory:

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/:id`
- `PATCH /v1/pvz/:id`
- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/:id/approve`
- `POST /v1/moderation/pvz/:id/reject`
- `POST /v1/moderation/pvz/:id/block`

Matching:

- `POST /v1/vacancies`
- `GET /v1/vacancies`
- `GET /v1/vacancies/:id`
- `POST /v1/vacancies/:id/responses`
- `GET /v1/vacancies/:id/responses`
- `POST /v1/responses/:id/withdraw`
- `POST /v1/responses/:id/accept`
- `POST /v1/assignments/:id/cancel`
- `POST /v1/assignments/:id/complete`
- `POST /v1/assignments/:id/reviews`
- `GET /v1/reviews`
- `POST /v1/reviews/:id/replies`

Communication:

- `GET /v1/conversations`
- `GET /v1/conversations/:id/messages`
- `POST /v1/conversations/:id/messages`
- `POST /v1/conversations/:id/read`
- `GET /ws/conversations/:id`

## Maintenance Rule

Every HTTP API behavior change must update this document and the Insomnia collection in `api-gateway/insomnia/berusmenu-api.json`.
