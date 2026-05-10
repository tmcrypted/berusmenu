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

- `GET /v1/auth/session`
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
  "role": "worker"
}
```

Fields:

- `email`: required, email.
- `password`: required, minimum 8 chars.
- `full_name`: optional. Frontend MVP collects it after email verification during profile completion.
- `role`: optional, one of `worker`, `owner`, `moderator`; defaults to `worker`.
- `phone`: optional. Frontend MVP collects it after email verification during profile completion.

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

### `GET /v1/auth/session`

Returns the currently authenticated user from the access cookie without rotating refresh cookies.

Auth: access cookie required. Allowed before email verification.

Success `200`:

```json
{
  "id": "7",
  "email": "worker@example.com",
  "role": "USER_ROLE_WORKER",
  "full_name": "Worker User",
  "created_at": "2026-05-06T13:00:00Z",
  "updated_at": "2026-05-06T13:00:00Z"
}
```

If the user is already verified, the response also includes `email_verified_at`.

Frontend note: Next.js uses this read-only endpoint for server-side first-paint auth bootstrap so the header and guarded routes render from the correct auth state without rotating refresh tokens.

Possible errors: `401`, `404`.

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

Frontend note: after successful email verification, `GET /v1/me` is used to decide whether the user is already profile-complete or should be redirected to a role-specific completion screen.

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

Frontend note: for the current MVP, `PATCH /v1/me` is the primary endpoint for post-verify profile completion.

Success `200`: same shape as `GET /v1/me`.

Possible errors: `400`, `401`, `403`, `404`.

## Inventory

All inventory routes require access cookie auth and verified email. Unsafe methods require `X-CSRF-Token`.

Inventory business rules live in `inventory-service`; gateway only maps HTTP to gRPC.

### `POST /v1/pvz`

Creates a PVZ for the current owner with status `pending` and appends a `submitted` moderation event.

Auth: owner.

Request:

```json
{
  "city": "Moscow",
  "address": "Tverskaya 1",
  "company_name": "Beru",
  "contact_phone": "+79990001122",
  "contact_telegram": "@beru_pvz",
  "description": "Main pickup point",
  "lat": 55.7558,
  "lng": 37.6173
}
```

Required fields: `city`, `address`, `company_name`, `contact_phone`.

Success `201`: PVZ object.

Possible errors: `400`, `401`, `403`, `500`.

### `GET /v1/pvz`

Returns the current owner's PVZ list.

Auth: owner.

Query:

- `page_size`: optional, default 20, max 100.
- `page_token`: optional numeric offset token.

Success `200`:

```json
{
  "items": [],
  "next_page_token": "20"
}
```

Possible errors: `400`, `401`, `403`, `500`.

### `GET /v1/pvz/{id}`

Returns a PVZ visible to the current user. Owners can read their own PVZ; moderators can inspect PVZ for moderation.

Success `200`: PVZ object.

Possible errors: `400`, `401`, `403`, `404`.

### `PATCH /v1/pvz/{id}`

Updates owner PVZ data, sets status back to `pending`, clears prior moderation decision fields, and appends a `submitted` moderation event.

Auth: owner. The owner can update only their own PVZ.

Request: same fields as create, all optional, but at least one field is required.

Temporary matching dependency: before mutating data, `inventory-service` calls `matching-service.HasActiveVacancyForPvz`. If matching reports an active vacancy or the RPC is unavailable, the PVZ is not changed.

Success `200`: PVZ object.

Possible errors: `400`, `401`, `403`, `404`, `412`, `503`.

### `GET /v1/moderation/pvz`

Returns pending PVZ for moderator review.

Auth: moderator.

Query: same pagination fields as `GET /v1/pvz`.

Success `200`:

```json
{
  "items": [],
  "next_page_token": "20"
}
```

Possible errors: `400`, `401`, `403`, `500`.

### `POST /v1/moderation/pvz/{id}/approve`

Approves a PVZ and appends an `approved` moderation event.

Auth: moderator.

Request:

```json
{
  "comment": "Documents checked"
}
```

`comment` is optional.

Success `200`: PVZ object.

Possible errors: `400`, `401`, `403`, `404`.

### `POST /v1/moderation/pvz/{id}/reject`

Rejects a PVZ, stores `rejection_reason`, and appends a `rejected` moderation event.

Auth: moderator.

Request:

```json
{
  "reason": "Invalid address",
  "comment": "Owner should update legal address"
}
```

`reason` is required; `comment` is optional.

Success `200`: PVZ object.

Possible errors: `400`, `401`, `403`, `404`.

### `POST /v1/moderation/pvz/{id}/block`

Blocks a PVZ and appends a `blocked` moderation event.

Auth: moderator.

Request:

```json
{
  "comment": "Fraud report confirmed"
}
```

`comment` is optional.

Success `200`: PVZ object.

Possible errors: `400`, `401`, `403`, `404`.

### `POST /v1/moderation/users/{id}/block-and-delete-pvz`

Temporarily blocks an owner account, deletes that owner's pending PVZ requests, and moves the owner's remaining PVZ to `blocked`.

Auth: moderator.

The gateway orchestrates two owning services: `auth-service.BlockUser` sets `blocked_until`, then `inventory-service.DeleteOwnerPvzRequests` cleans only `inventory.*` rows. If the second step fails after the user was blocked, the operation can be retried with the same user id.

Request:

```json
{
  "reason": "Spam PVZ submissions",
  "block_duration_minutes": 1440,
  "comment": "Bulk spam cleanup"
}
```

`reason` and `block_duration_minutes` are required; `comment` is optional.

Success `200`:

```json
{
  "blocked_user_id": "42",
  "blocked_until": "2026-05-10T12:00:00Z",
  "deleted_count": "3",
  "blocked_count": "2"
}
```

Possible errors: `400`, `401`, `403`, `404`, `412`, `500`.

## Matching

All matching routes require access cookie auth and verified email. Unsafe methods require `X-CSRF-Token`.

Vacancy shift windows use ISO 8601 timestamps:

```json
{
  "shift_start_at": "2026-05-10T09:00:00+03:00",
  "shift_end_at": "2026-05-10T21:00:00+03:00"
}
```

`GET /v1/vacancies` is role-aware:

- worker sees published vacancies;
- owner sees own vacancies;
- moderator sees all vacancies.

Supported query parameters for `GET /v1/vacancies`:

- `status`: optional, one of `published`, `assigned`, `completed`, `cancelled`; invalid values return `400 invalid_argument`.
- `pvz_id`: optional positive id.
- `shift_start_from`: optional ISO 8601 timestamp.
- `shift_start_to`: optional ISO 8601 timestamp.
- `page_size`: optional, default service limit, max service limit.
- `page_token`: optional numeric offset token.

Implemented public matching routes:

- `POST /v1/vacancies`
- `GET /v1/vacancies`
- `GET /v1/vacancies/:id`
- `POST /v1/vacancies/:id/responses`
- `GET /v1/vacancies/:id/responses`
- `GET /v1/vacancies/:id/assignment`
- `GET /v1/responses/me`
- `POST /v1/responses/:id/withdraw`
- `POST /v1/responses/:id/accept`
- `GET /v1/assignments/me`
- `GET /v1/assignments/:id`
- `POST /v1/assignments/:id/cancel`
- `POST /v1/assignments/:id/complete`
- `POST /v1/assignments/:id/reviews`
- `PATCH /v1/reviews/:id`
- `GET /v1/reviews`
- `POST /v1/reviews/:id/replies`
- `GET /v1/reviews/:id/replies`
- `PATCH /v1/review-replies/:id`
- `GET /v1/users/:id/rating`

### `POST /v1/moderation/matching/responses/{id}/accept`

Moderator-only dispute action. Accepts a pending response, creates the assignment, and writes `matching.admin_events` in the same matching transaction.

Request:

```json
{
  "reason": "Moderator dispute resolution",
  "comment": "Owner is unavailable"
}
```

`reason` is required. `comment` is accepted by the HTTP DTO for future moderation-service extraction.

### `POST /v1/moderation/matching/assignments/{id}/cancel`

Moderator-only dispute action. Cancels an active assignment and writes `matching.admin_events` in the same matching transaction.

Request:

```json
{
  "reason": "Moderator dispute resolution",
  "cancel_reason": "Worker is unavailable",
  "comment": "Evidence checked"
}
```

`reason` is required. `cancel_reason` is stored on the assignment.

### `POST /v1/moderation/matching/assignments/{id}/complete`

Moderator-only dispute action. Completes an active assignment and writes `matching.admin_events` in the same matching transaction.

Request:

```json
{
  "reason": "Moderator dispute resolution",
  "comment": "Shift completion confirmed"
}
```

`reason` is required.

Moderator matching routes are exposed under `/v1/moderation/matching/...` for future extraction. MVP execution remains in `matching-service`, and gateway applies a moderator-only gate before calling gRPC.

## Placeholder Endpoints

These routes are registered in the gateway but currently return `not_implemented` until `communication-service` is implemented.

Communication:

- `GET /v1/conversations`
- `GET /v1/conversations/:id/messages`
- `POST /v1/conversations/:id/messages`
- `POST /v1/conversations/:id/read`
- `GET /ws/conversations/:id`

## Maintenance Rule

Every HTTP API behavior change must update this document and the Insomnia collection in `api-gateway/insomnia/berusmenu-api.json`.
