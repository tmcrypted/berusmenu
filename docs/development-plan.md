# Development Plan

## Phase 1. Contracts and persistence

Status: mostly done for foundation contracts; domain services still need implementation-level refinement.

1. Freeze REST and gRPC outlines for implemented auth and notification flows.
2. Freeze state transitions and permissions.
3. Create protobuf contracts in `proto`.
4. Prepare migrations for:
   - `auth`
   - `inventory`
   - `matching`
   - `communication`

## Phase 2. Notification and auth

Status: foundation implemented.

1. `notification-service` module exists inside the monorepo.
2. Kafka consumer is implemented.
3. Email delivery is currently a log/stub sender abstraction; Mailganer integration remains future work.
4. `auth-service` implements:
   - register
   - login
   - atomic refresh rotation/logout
   - resend verification code
   - verify email
   - JWT restrictions for unverified users
5. `auth-service` business logic is split into `internal/service`, with gRPC server kept as a thin adapter.
6. `api-gateway` uses HttpOnly cookie auth with CSRF validation for web clients.
7. `notification-service` consumes both MVP notification topics with retry-aware commit policy.

## Phase 3. Inventory and matching

Status: pending.

1. Implement `inventory-service`.
2. Implement `matching-service`.
3. Implement `communication-service`.
4. After `matching-service` is implemented, replace the temporary inventory active-vacancy checker test path with real `HasActiveVacancyForPvz` integration coverage for `published` and `assigned` vacancies.

## Phase 4. Edge and UI

Status: in progress.

1. `api-gateway` has auth/profile HTTP routes and placeholder domain routes.
2. HTTP API docs live in `api-gateway/docs/http-api.md`.
3. Manual API testing collection lives in `api-gateway/insomnia/berusmenu-api.json`.
4. Extend `api-gateway` REST -> gRPC mapping for inventory, matching, and communication.
5. Implement frontend auth flow with verify-code screen.
6. Implement owner PVZ flow.
7. Implement owner vacancy flow.
8. Implement worker response flow.
9. Implement chat UI.
10. Implement reviews UI.

## Phase 5. Stabilization

1. Integration tests for main business paths.
2. CI/CD wiring for the monorepo.
3. Deployment notes in `infrastructure`.
