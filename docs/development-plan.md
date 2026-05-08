# Development Plan

This file tracks the current roadmap after the auth, notification and inventory slices.

## Phase 1. Foundation Contracts

Status: done enough for current MVP backend slices.

- Protobuf modules exist for `auth`, `inventory`, `matching`, `communication`, `events`, `common`.
- `go.work` joins implemented Go modules.
- HTTP API documentation and Insomnia collection exist under `api-gateway`.
- Product rules are split across architecture/system-design/permissions/state-transition docs.

## Phase 2. Auth And Notification

Status: implemented.

- `auth-service` implements register/login/refresh/logout.
- JWT validation is owned by `auth-service`.
- Email verification code storage is Redis-backed.
- Resend cooldown and failed-attempt lockout are implemented.
- Current profile read/update is implemented.
- Moderator timed block is implemented via `auth.users.blocked_until`.
- `api-gateway` uses HttpOnly cookies and CSRF protection for web clients.
- `notification-service` consumes:
  - `auth.email_verification_requested`
  - `matching.response_created`
- Email delivery is a local log/stub sender. Production provider remains future work.

## Phase 3. Inventory

Status: backend and frontend slice implemented.

- `inventory-service` implements:
  - create PVZ
  - get/list owner PVZ
  - update PVZ
  - check approved PVZ for future vacancy creation
  - list pending PVZ
  - approve/reject/block PVZ
  - owner spam cleanup for moderator flow
- `api-gateway` exposes inventory HTTP routes.
- `frontend` contains owner PVZ pages and moderator moderation panel.
- Temporary matching dependency exists through `ActiveVacancyChecker`.

Remaining inventory-adjacent work:

- replace temporary active-vacancy path with real `matching-service.HasActiveVacancyForPvz`;
- add end-to-end test proving PVZ update is denied for `published` and `assigned` vacancies;
- manually verify owner/moderator flows in Insomnia and browser.

## Phase 4. Matching

Status: planned / next major backend service.

Implement `matching-service`:

1. migrations for `matching.vacancies`, `matching.responses`, `matching.assignments`, `matching.reviews`, `matching.review_replies`;
2. vacancy create/list/get flows;
3. owner vacancy list;
4. response create/withdraw/list/accept;
5. assignment create/cancel/complete;
6. reviews and review replies;
7. `HasActiveVacancyForPvz`;
8. `CheckResponseChatWritable`;
9. Kafka publish for `matching.response_created`.

## Phase 5. Communication

Status: planned.

Implement `communication-service`:

1. create conversation for response;
2. list conversations;
3. list messages;
4. send message;
5. mark as read;
6. Redis-backed presence/typing/active websocket state;
7. gateway websocket edge.

## Phase 6. Frontend Domain Flows

Status: auth/profile/inventory done, matching/communication pending.

Implemented:

- public home;
- register/login/logout;
- email verification;
- post-verification profile completion;
- owner PVZ list/create/edit;
- moderator PVZ queue and moderation actions.

Pending:

- owner vacancy flow;
- worker vacancy browsing and response flow;
- assignment status actions;
- chat UI;
- reviews and replies UI.

## Phase 7. Stabilization

- Integration tests for main business paths.
- CI for Go and frontend checks.
- Production deployment notes.
- Production email provider integration.
- Observability baseline only after core flows are stable.
