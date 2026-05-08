# API Gateway

`api-gateway` is the public HTTP entrypoint for Berusmenu.

It owns:

- public REST shape;
- HttpOnly cookie auth transport;
- CSRF checks for authenticated unsafe methods;
- credentialed CORS for the frontend origin;
- verified-user gate;
- REST -> gRPC mapping;
- gRPC identity metadata forwarding.

Business rules stay in owning services.

## Current Scope

Implemented:

- health endpoints;
- auth endpoints;
- session bootstrap endpoint;
- current profile endpoints;
- inventory owner routes;
- PVZ moderation routes;
- moderator owner spam cleanup route;
- placeholder route groups for future matching and communication flows;
- gRPC clients to downstream services.

## HTTP Routes

Health:

- `GET /health/live`
- `GET /health/ready`

Auth/profile:

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `GET /v1/auth/session`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

Inventory:

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/{id}`
- `PATCH /v1/pvz/{id}`
- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/{id}/approve`
- `POST /v1/moderation/pvz/{id}/reject`
- `POST /v1/moderation/pvz/{id}/block`
- `POST /v1/moderation/users/{id}/block-and-delete-pvz`

## Web Auth Transport

The public web API does not use `Authorization: Bearer`.

- `berusmenu_access`: HttpOnly access JWT cookie.
- `berusmenu_refresh`: HttpOnly refresh token cookie scoped to auth routes.
- `berusmenu_csrf`: readable CSRF cookie.
- `X-CSRF-Token`: required for authenticated `POST`, `PATCH`, `PUT`, `DELETE`.

Gateway validates the access cookie through `auth-service` and forwards identity downstream as gRPC metadata:

- `x-user-id`
- `x-user-role`
- `x-user-email`
- `x-user-email-verified`

## Unverified User Policy

JWT may exist before email verification.

Allowed while unverified:

- logout;
- resend verification code;
- verify email;
- read-only session bootstrap.

Verified-only:

- `GET /v1/me`;
- `PATCH /v1/me`;
- all inventory/matching/communication domain routes.

## Env

| Variable | Default |
| --- | --- |
| `HTTP_ADDR` | `:8080` |
| `AUTH_GRPC_ADDR` | `localhost:50051` |
| `INVENTORY_GRPC_ADDR` | `localhost:50052` |
| `MATCHING_GRPC_ADDR` | `localhost:50053` |
| `COMMUNICATION_GRPC_ADDR` | `localhost:50054` |
| `REQUEST_TIMEOUT` | `5s` |
| `APP_ENV` | `local` |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` |
| `COOKIE_SECURE` | `false` |
| `COOKIE_DOMAIN` | empty |
| `ACCESS_COOKIE_NAME` | `berusmenu_access` |
| `REFRESH_COOKIE_NAME` | `berusmenu_refresh` |
| `CSRF_COOKIE_NAME` | `berusmenu_csrf` |
| `CSRF_HEADER_NAME` | `X-CSRF-Token` |
| `REFRESH_COOKIE_TTL` | `720h` |

## Local Run

From service root:

```powershell
cd C:\Users\tmcrypted\Desktop\berusmenu\api-gateway
$env:AUTH_GRPC_ADDR="localhost:50051"
$env:INVENTORY_GRPC_ADDR="localhost:50052"
go run .\cmd\api-gateway
```

## Tests

```powershell
go test ./...
```

## API Documentation

- HTTP API reference: [`docs/http-api.md`](docs/http-api.md)
- Insomnia import collection: [`insomnia/berusmenu-api.json`](insomnia/berusmenu-api.json)

Every HTTP API behavior change must update both files together with the code.
