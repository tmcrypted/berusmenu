# System Design Berusmenu MVP

This document is the current design baseline. More detailed slices live in:

- [Sequence Diagrams](sequence-diagrams.md)
- [State Transitions](state-transitions.md)
- [Permissions](permissions.md)
- [Interfaces](interfaces.md)
- [Events And Cache](events-and-cache.md)
- [Runtime Topology](runtime-topology.md)
- [Development Plan](development-plan.md)

## Product Goal

Help a PVZ owner quickly publish an urgent shift, receive worker responses, choose an assignee, chat, complete/cancel the assignment and collect reviews. MVP optimizes for a short vacancy-to-assignment path and simple operations.

## Current Implementation Status

Implemented:

- auth backend;
- notification worker;
- inventory backend;
- auth/profile frontend;
- owner PVZ frontend;
- moderator PVZ frontend;
- gateway routes for auth/profile/inventory.

Planned:

- matching backend/frontend;
- communication backend/frontend;
- reviews UI;
- production email provider.

## Must-Have Scope

- email/password auth;
- HttpOnly cookie web auth with CSRF;
- email verification code;
- profile completion after verification;
- PVZ registration and moderation;
- moderator timed spam block for abusive owners;
- vacancy creation;
- worker response;
- chat inside response flow;
- assignment selection;
- assignment cancellation/completion;
- reviews and review replies;
- email notifications for verification and new response.

## Core Components

- `frontend`
- `api-gateway`
- `auth-service`
- `inventory-service`
- `matching-service`
- `communication-service`
- `notification-service`
- `postgres`
- `redis`
- `kafka`

## Persisted Data

### `auth`

- `users`
- `refresh_sessions`
- `outbox_events`

`users.blocked_until` supports temporary moderator blocks.

### `inventory`

- `pvz`
- `pvz_moderation_events`

### `matching`

Planned:

- `vacancies`
- `responses`
- `assignments`
- `reviews`
- `review_replies`

### `communication`

Planned:

- `conversations`
- `messages`
- `message_reads`

### `notification`

- `processed_events`

## Core Rules

- One database, separate schema per service.
- FK only inside a service schema.
- Cross-domain integrity is validated through owning services.
- Gateway stays thin.
- User receives JWT after registration, but verified-only flows are blocked until email verification.
- Auth tokens are stored in HttpOnly cookies for web clients.
- PVZ update mutates `inventory.pvz`, sets it back to `pending`, and requires repeated moderation.
- PVZ update is forbidden when active vacancy exists with status `published` or `assigned`.
- New vacancy cannot be created while PVZ is not `approved`.
- Moderator spam cleanup temporarily blocks owner in `auth-service`, deletes pending owner PVZ requests, and sets remaining owner PVZ to `blocked`.
- Conversation is created immediately after response.
- Repeat response after `withdrawn` is forbidden.
- Old vacancy is not reused after `assignment.cancelled`.
- Reviews are allowed after `completed` or `cancelled`.
- Replies belong only to root reviews and have one nesting level.
- Chat is read-only after assignment finalization.

## Async Rules

Kafka is used only for notification events:

- `auth.email_verification_requested`
- `matching.response_created`

Consumer:

- `notification-service`

Delivery:

- producers persist notification events to service-owned outbox tables before Kafka publish;
- notification events include `event_id` for idempotent processing;
- email only in MVP;
- local log/stub sender now;
- production provider later.

## Non-Goals

- payments;
- advanced analytics;
- recommendation engine;
- CQRS/event sourcing;
- multi-channel notifications beyond email;
- full observability stack before core flows are stable.
