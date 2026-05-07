# REST и gRPC Outline

## Public REST via API Gateway

Auth transport:

- Web clients authenticate with `berusmenu_access` HttpOnly cookie.
- Refresh uses `berusmenu_refresh` HttpOnly cookie.
- Authenticated unsafe methods require `X-CSRF-Token` matching `berusmenu_csrf`.
- Gateway forwards verified identity to gRPC services through metadata.

### Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me`
- `PATCH /v1/me`

### PVZ

- `POST /v1/pvz`
- `GET /v1/pvz`
- `GET /v1/pvz/{id}`
- `PATCH /v1/pvz/{id}`

### Moderation

- `GET /v1/moderation/pvz`
- `POST /v1/moderation/pvz/{id}/approve`
- `POST /v1/moderation/pvz/{id}/reject`
- `POST /v1/moderation/pvz/{id}/block`

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

## gRPC outline

### AuthService

- `Register`
- `Login`
- `Refresh`
- `Logout`
- `ValidateToken`
- `GetUser`
- `UpdateProfile`
- `ResendVerificationCode`
- `VerifyEmail`

### InventoryService

- `CreatePvz`
- `GetPvz`
- `ListOwnerPvz`
- `UpdatePvz`
- `CheckPvzForVacancy`
- `ListPendingPvz`
- `ApprovePvz`
- `RejectPvz`
- `BlockPvz`

### MatchingService

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

- `CreateConversationForResponse`
- `ListConversations`
- `ListMessages`
- `SendMessage`
- `MarkAsRead`

### NotificationService

No public API in MVP.

Only Kafka consumer handlers:

- handle `auth.email_verification_requested`
- handle `matching.response_created`
