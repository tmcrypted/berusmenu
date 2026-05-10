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

### Matching Flow

Vacancies:

- `POST /v1/vacancies`
- `GET /v1/vacancies`
- `GET /v1/vacancies/{id}`

Responses:

- `POST /v1/vacancies/{id}/responses`
- `GET /v1/vacancies/{id}/responses`
- `GET /v1/vacancies/{id}/assignment`
- `GET /v1/responses/me`
- `POST /v1/responses/{id}/withdraw`
- `POST /v1/responses/{id}/accept`

Assignments:

- `GET /v1/assignments/me`
- `GET /v1/assignments/{id}`
- `POST /v1/assignments/{id}/cancel`
- `POST /v1/assignments/{id}/complete`

Reviews:

- `POST /v1/assignments/{id}/reviews`
- `GET /v1/reviews`
- `PATCH /v1/reviews/{id}`
- `POST /v1/reviews/{id}/replies`
- `GET /v1/reviews/{id}/replies`
- `PATCH /v1/review-replies/{id}`
- `GET /v1/users/{id}/rating`

Matching moderation:

- `GET /v1/moderation/matching/vacancies`
- `POST /v1/moderation/matching/responses/{id}/accept`
- `POST /v1/moderation/matching/assignments/{id}/cancel`
- `POST /v1/moderation/matching/assignments/{id}/complete`

## Planned REST

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

Implemented:

- `CreateVacancy`
- `ListVacancies`
- `ListOwnerVacancies`
- `GetVacancy`
- `CreateResponse`
- `WithdrawResponse`
- `ListResponsesForVacancy`
- `ListMyResponses`
- `AcceptResponse`
- `GetAssignment`
- `ListMyAssignments`
- `GetAssignmentForVacancy`
- `CancelAssignment`
- `CompleteAssignment`
- `CreateReview`
- `UpdateReview`
- `CreateReviewReply`
- `UpdateReviewReply`
- `ListReviews`
- `ListReviewReplies`
- `GetUserRating`
- `HasActiveVacancyForPvz`
- `CheckResponseChatWritable`

Vacancy create/list uses full timestamp fields:

- `shift_start_at`
- `shift_end_at`
- list filters `shift_start_from` / `shift_start_to`

`GET /v1/vacancies` is role-aware:

- worker sees published vacancies;
- owner sees own vacancies;
- moderator sees all vacancies.

Moderator matching HTTP routes are exposed under `/v1/moderation/matching/...` for future extraction. Gateway applies a moderator-only gate, and MVP execution remains inside `matching-service`.

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
