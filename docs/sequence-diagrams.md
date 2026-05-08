# Sequence Diagrams

## 1. Register + send verification code

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant GW as API Gateway
    participant AU as Auth Service
    participant PG as PostgreSQL
    participant R as Redis
    participant K as Kafka
    participant NS as Notification Service
    participant ES as Email Sender

    U->>FE: Submit registration form
    FE->>GW: POST /v1/auth/register
    GW->>AU: Register(email, password, role)
    AU->>PG: Check email uniqueness
    PG-->>AU: Email is free
    AU->>PG: Create user
    PG-->>AU: user_id
    AU->>PG: Create refresh session
    AU->>R: SET email_verify:{user_id}=code TTL 10m
    AU->>K: Publish auth.email_verification_requested
    AU-->>GW: user + access_token + refresh_token
    GW-->>FE: 201 Created + HttpOnly auth cookies + CSRF cookie
    K-->>NS: auth.email_verification_requested
    NS->>ES: Log/send verification email
```

## 2. Resend verification code

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant GW as API Gateway
    participant AU as Auth Service
    participant R as Redis
    participant K as Kafka
    participant NS as Notification Service
    participant ES as Email Sender

    U->>FE: Click resend code
    FE->>GW: POST /v1/auth/resend-code
    GW->>AU: ResendVerificationCode
    AU->>R: Check email_verify_resend:{user_id}
    AU->>R: Read current code from email_verify:{user_id}
    AU->>K: Publish auth.email_verification_requested
    AU->>R: SET email_verify_resend:{user_id} TTL 60s
    AU-->>GW: accepted
    GW-->>FE: 202 Accepted
    K-->>NS: auth.email_verification_requested
    NS->>ES: Log/send same active code
```

## 3. Verify email

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant GW as API Gateway
    participant AU as Auth Service
    participant R as Redis
    participant PG as PostgreSQL

    U->>FE: Enter verification code
    FE->>GW: POST /v1/auth/verify-email
    GW->>AU: VerifyEmail(code)
    AU->>R: Check lockout and attempts
    AU->>R: GET email_verify:{user_id}
    AU->>AU: Compare code
    alt valid
        AU->>PG: UPDATE users.email_verified_at
        AU->>R: DELETE email_verify:{user_id}
        AU->>R: DELETE email_verify_attempts:{user_id}
        AU-->>GW: success
        GW-->>FE: 200 OK
    else invalid
        AU->>R: Increment email_verify_attempts:{user_id}
        AU-->>GW: error
        GW-->>FE: 400/429
    end
```

## 4. Login verified user

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant GW as API Gateway
    participant AU as Auth Service
    participant PG as PostgreSQL

    U->>FE: Submit login form
    FE->>GW: POST /v1/auth/login
    GW->>AU: Login(email, password)
    AU->>PG: Find user by email
    AU->>AU: Verify password
    AU->>PG: Create refresh session
    AU-->>GW: user + access_token + refresh_token
    GW-->>FE: 200 OK + HttpOnly auth cookies + CSRF cookie
```

## 5. Create PVZ

```mermaid
sequenceDiagram
    participant Owner
    participant FE as Frontend
    participant GW as API Gateway
    participant INV as Inventory Service
    participant PG as PostgreSQL

    Owner->>FE: Create PVZ
    FE->>GW: POST /v1/pvz
    GW->>INV: CreatePvz
    INV->>PG: Insert inventory.pvz with status pending
    INV->>PG: Insert moderation event submitted
    INV-->>GW: PVZ created
    GW-->>FE: 201 Created
```

## 6. Approve / Reject PVZ

```mermaid
sequenceDiagram
    participant Mod as Moderator
    participant FE as Frontend
    participant GW as API Gateway
    participant INV as Inventory Service
    participant PG as PostgreSQL

    Mod->>FE: Approve or reject PVZ
    FE->>GW: POST /v1/moderation/pvz/{id}/approve or /reject
    GW->>INV: ApprovePvz or RejectPvz
    INV->>PG: Update pvz status
    INV->>PG: Insert moderation event
    INV-->>GW: done
    GW-->>FE: 200 OK
```

## 7. Update PVZ and move back to pending

```mermaid
sequenceDiagram
    participant Owner
    participant FE as Frontend
    participant GW as API Gateway
    participant INV as Inventory Service
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    Owner->>FE: Update PVZ
    FE->>GW: PATCH /v1/pvz/{id}
    GW->>INV: UpdatePvz
    INV->>MATCH: HasActiveVacancyForPvz(pvz_id)
    MATCH-->>INV: active vacancy flag
    alt active vacancy exists
        INV-->>GW: reject update
        GW-->>FE: 412 Failed Precondition or 503 Unavailable
    else no active vacancy
        INV->>PG: Update PVZ data and set status pending
        INV->>PG: Insert moderation event submitted
        INV-->>GW: accepted
        GW-->>FE: 200 OK
    end
```

## 8. Create vacancy

```mermaid
sequenceDiagram
    participant Owner
    participant FE as Frontend
    participant GW as API Gateway
    participant INV as Inventory Service
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    Owner->>FE: Create vacancy
    FE->>GW: POST /v1/vacancies
    GW->>MATCH: CreateVacancy
    MATCH->>INV: CheckPvzForVacancy(pvz_id)
    INV-->>MATCH: approved owner PVZ snapshot
    MATCH->>PG: Insert matching.vacancies
    MATCH-->>GW: vacancy created
    GW-->>FE: 201 Created
```

## 8a. Moderator spam cleanup for owner

```mermaid
sequenceDiagram
    participant Mod as Moderator
    participant FE as Frontend
    participant GW as API Gateway
    participant AU as Auth Service
    participant INV as Inventory Service
    participant APG as auth.users
    participant IPG as inventory.pvz

    Mod->>FE: Block spam owner for N minutes
    FE->>GW: POST /v1/moderation/users/{owner_id}/block-and-delete-pvz
    GW->>AU: BlockUser(owner_id, blocked_until, reason)
    AU->>APG: status=blocked, blocked_until=now+N
    AU-->>GW: user blocked
    GW->>INV: DeleteOwnerPvzRequests(owner_id, comment)
    INV->>IPG: Delete pending owner PVZ rows
    INV->>IPG: Set remaining owner PVZ to blocked
    INV->>IPG: Insert blocked moderation events
    INV-->>GW: deleted_count, blocked_count
    GW-->>FE: 200 OK
```

## 9. Create response + create conversation + notify owner

```mermaid
sequenceDiagram
    participant Worker
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant COMM as Communication Service
    participant PG as PostgreSQL
    participant K as Kafka
    participant NS as Notification Service
    participant ES as Email Sender

    Worker->>FE: Respond to vacancy
    FE->>GW: POST /v1/vacancies/{id}/responses
    GW->>MATCH: CreateResponse
    MATCH->>PG: Insert matching.responses
    MATCH->>COMM: CreateConversationForResponse
    COMM->>PG: Insert communication.conversations
    MATCH->>K: Publish matching.response_created
    MATCH-->>GW: response created
    GW-->>FE: 201 Created
    K-->>NS: matching.response_created
    NS->>ES: Log/send owner email notification
```

## 10. Send message

```mermaid
sequenceDiagram
    participant User
    participant FE as Frontend
    participant GW as API Gateway
    participant COMM as Communication Service
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    User->>FE: Send message
    FE->>GW: WS/HTTP message send
    GW->>COMM: SendMessage
    COMM->>MATCH: CheckResponseChatWritable(response_id)
    MATCH-->>COMM: writable flag
    COMM->>PG: Insert communication.messages
    COMM->>PG: Update conversations.last_message_at
    COMM-->>GW: message stored
    GW-->>FE: delivered
```

## 11. Accept response -> assignment

```mermaid
sequenceDiagram
    participant Owner
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    Owner->>FE: Accept response
    FE->>GW: POST /v1/responses/{id}/accept
    GW->>MATCH: AcceptResponse
    MATCH->>PG: Update chosen response to accepted
    MATCH->>PG: Update other responses to rejected
    MATCH->>PG: Insert assignment
    MATCH->>PG: Update vacancy status to assigned
    MATCH-->>GW: assignment created
    GW-->>FE: 200 OK
```

## 12. Cancel assignment

```mermaid
sequenceDiagram
    participant User
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    User->>FE: Cancel assignment
    FE->>GW: POST /v1/assignments/{id}/cancel
    GW->>MATCH: CancelAssignment
    MATCH->>PG: Update assignment to cancelled
    MATCH->>PG: Update vacancy to cancelled
    MATCH-->>GW: cancelled
    GW-->>FE: 200 OK
```

## 13. Complete assignment

```mermaid
sequenceDiagram
    participant Owner
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    Owner->>FE: Complete assignment
    FE->>GW: POST /v1/assignments/{id}/complete
    GW->>MATCH: CompleteAssignment
    MATCH->>PG: Update assignment to completed
    MATCH->>PG: Update vacancy to completed
    MATCH-->>GW: completed
    GW-->>FE: 200 OK
```

## 14. Create review

```mermaid
sequenceDiagram
    participant User
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    User->>FE: Create review
    FE->>GW: POST /v1/assignments/{id}/reviews
    GW->>MATCH: CreateReview
    MATCH->>PG: Check assignment final status
    MATCH->>PG: Insert review
    MATCH-->>GW: review created
    GW-->>FE: 201 Created
```

## 15. Create review reply

```mermaid
sequenceDiagram
    participant User
    participant FE as Frontend
    participant GW as API Gateway
    participant MATCH as Matching Service
    participant PG as PostgreSQL

    User->>FE: Reply to review
    FE->>GW: POST /v1/reviews/{id}/replies
    GW->>MATCH: CreateReviewReply
    MATCH->>PG: Check review and assignment participants
    MATCH->>PG: Insert review reply
    MATCH-->>GW: reply created
    GW-->>FE: 201 Created
```
