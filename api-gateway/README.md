# API Gateway

REST/JSON BFF для Berusmenu.

Gateway принимает запросы от frontend, валидирует auth context и маршрутизирует операции во внутренние gRPC сервисы.

## Ответственность

- public HTTP API
- auth middleware
- restriction logic for unverified users
- REST -> gRPC mapping
- websocket edge for chat

## Current scope

- health endpoints
- auth endpoints
- current user profile
- skeleton route groups for inventory, matching and communication
- gRPC clients to downstream services

## Endpoints

### Health

- `GET /health/live`
- `GET /health/ready`
- `GET /openapi.json`
- `GET /swagger`

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

## Unverified user policy

JWT can already exist before email verification.

Gateway must allow only whitelist endpoints for unverified users:

- logout
- resend verification code
- verify email
- `me` is not available for unverified users

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

## Local run

```bash
go run ./cmd/api-gateway
```

## Tests

```bash
go test ./...
```

## Notes

- `gitlab.com/berusmenu/proto` is resolved locally through `replace => ../proto`
- gateway forwards identity through gRPC metadata
- gateway should stay thin; domain rules stay in owning services
- built-in OpenAPI JSON is served from `/openapi.json`
- built-in docs UI is served from `/swagger`
