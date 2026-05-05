# Development Plan

## Phase 1. Contracts and persistence

1. Freeze REST and gRPC outlines.
2. Freeze state transitions and permissions.
3. Create protobuf contracts in `proto`.
4. Prepare goose migrations for:
   - `auth`
   - `inventory`
   - `matching`
   - `communication`

## Phase 2. Notification and auth

1. Create `notification-service` repo skeleton.
2. Implement Kafka consumer and Mailganer sender.
3. Implement `auth-service`:
   - register
   - login
   - refresh/logout
   - resend verification code
   - verify email
   - JWT restrictions for unverified users

## Phase 3. Inventory and matching

1. Implement `inventory-service`.
2. Implement `matching-service`.
3. Implement `communication-service`.

## Phase 4. Edge and UI

1. Extend `api-gateway`.
2. Implement frontend auth flow with verify-code screen.
3. Implement owner PVZ flow.
4. Implement owner vacancy flow.
5. Implement worker response flow.
6. Implement chat UI.
7. Implement reviews UI.

## Phase 5. Stabilization

1. Integration tests for main business paths.
2. CI/CD wiring in GitLab.
3. Deployment notes in `infrastructure`.
