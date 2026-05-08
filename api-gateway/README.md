# API Gateway

REST/JSON BFF для Berusmenu.

Gateway принимает запросы от frontend, валидирует auth context и маршрутизирует операции во внутренние gRPC сервисы.

## Ответственность

- public HTTP API
- HttpOnly cookie auth middleware
- CSRF validation for authenticated unsafe methods
- credentialed CORS for configured frontend origins
- restriction logic for unverified users
- REST -> gRPC mapping
- websocket edge for chat

## Current scope

- health endpoints
- auth endpoints
- current user profile
- inventory PVZ and moderation REST -> gRPC routes
- skeleton route groups for matching and communication
- gRPC clients to downstream services

## Endpoints

### Health

- `GET /health/live`
- `GET /health/ready`

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

### Inventory

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/{id}`
- `PATCH /v1/pvz/{id}`
- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/{id}/approve`
- `POST /v1/moderation/pvz/{id}/reject`
- `POST /v1/moderation/pvz/{id}/block`

## Unverified user policy

JWT can already exist before email verification.

Gateway must allow only whitelist endpoints for unverified users:

- logout
- resend verification code
- verify email
- `me` is not available for unverified users

## Web auth transport

The public web API does not use `Authorization: Bearer`.

- `berusmenu_access`: HttpOnly access JWT cookie.
- `berusmenu_refresh`: HttpOnly refresh token cookie scoped to auth routes.
- `berusmenu_csrf`: readable CSRF cookie.
- `X-CSRF-Token`: must match `berusmenu_csrf` for authenticated `POST`, `PATCH`, `PUT`, and `DELETE`.

Gateway validates the access cookie through `auth-service` and forwards identity downstream as gRPC metadata.

The Insomnia collection uses the built-in Cookie template tag for CSRF:

```text
{% cookie _.base_url, 'berusmenu_csrf' %}
```

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

## Local run

```bash
go run ./cmd/api-gateway
```

## Tests

```bash
go test ./...
```

## API documentation

- HTTP API reference: [`docs/http-api.md`](docs/http-api.md)
- Insomnia import collection: [`insomnia/berusmenu-api.json`](insomnia/berusmenu-api.json)

Every HTTP API behavior change must update both files together with the code.

## Notes

- `github.com/tmcrypted/berusmenu/proto` is resolved locally through `replace => ../proto`
- gateway forwards identity through gRPC metadata
- gateway should stay thin; domain rules stay in owning services
