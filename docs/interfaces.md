# REST And gRPC Interfaces

## Public REST via API Gateway

Web auth transport:

- `berusmenu_access` HttpOnly cookie carries the access JWT.
- `berusmenu_refresh` HttpOnly cookie carries the refresh token.
- `berusmenu_csrf` is readable by frontend.
- Authenticated unsafe methods require `X-CSRF-Token` equal to `berusmenu_csrf`.
- Gateway validates cookies through `auth-service` and forwards identity downstream through gRPC metadata.

## Implemented REST

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `GET /v1/auth/session`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

### PVZ Owner Flow

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/{id}`
- `PATCH /v1/pvz/{id}`

### PVZ Moderation Flow

- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/{id}/approve`
- `POST /v1/moderation/pvz/{id}/reject`
- `POST /v1/moderation/pvz/{id}/block`
- `POST /v1/moderation/users/{id}/block-and-delete-pvz`

The last endpoint orchestrates two owning services: timed user block in `auth-service`, then owner PVZ cleanup in `inventory-service`.

## Planned REST

### Vacancies

- `POST /v1/vacancies`
- `GET /v1/vacancies`
- `GET /v1/vacancies/{id}`

### Responses

- `POST /v1/vacancies/{id}/responses`
- `POST /v1/responses/{id}/withdraw`
- `POST /v1/responses/{id}/accept`

### Assignments

- `POST /v1/assignments/{id}/cancel`
- `POST /v1/assignments/{id}/complete`

### Reviews

- `POST /v1/assignments/{id}/reviews`
- `GET /v1/reviews`
- `POST /v1/reviews/{id}/replies`

### Chat

- `GET /v1/conversations`
- `GET /v1/conversations/{id}/messages`
- `POST /v1/conversations/{id}/read`
- `WS /ws/conversations/{id}`

## gRPC Services

### AuthService

Implemented:

- `Register`
- `Login`
- `Refresh`
- `Logout`
- `ValidateToken`
- `GetUser`
- `UpdateProfile`
- `ResendVerificationCode`
- `VerifyEmail`
- `BlockUser`

`BlockUser` requires moderator identity metadata and `blocked_until`.

### InventoryService

Implemented:

- `CreatePvz`
- `GetPvz`
- `ListOwnerPvz`
- `UpdatePvz`
- `CheckPvzForVacancy`
- `ListPendingPvz`
- `ApprovePvz`
- `RejectPvz`
- `BlockPvz`
- `DeleteOwnerPvzRequests`

`DeleteOwnerPvzRequests` currently deletes owner pending PVZ rows and blocks remaining owner PVZ rows. The RPC name is intentionally transport-level legacy; service behavior is documented in inventory docs.

### MatchingService

Planned:

- `CreateVacancy`
- `ListVacancies`
- `ListOwnerVacancies`
- `GetVacancy`
- `CreateResponse`
- `WithdrawResponse`
- `ListResponsesForVacancy`
- `AcceptResponse`
- `CancelAssignment`
- `CompleteAssignment`
- `CreateReview`
- `CreateReviewReply`
- `ListReviews`
- `HasActiveVacancyForPvz`
- `CheckResponseChatWritable`

### CommunicationService

Planned:

- `CreateConversationForResponse`
- `ListConversations`
- `ListMessages`
- `SendMessage`
- `MarkAsRead`

### NotificationService

No public API in MVP.

Kafka handlers:

- `auth.email_verification_requested`
- `matching.response_created`
