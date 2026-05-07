# Permissions

## Roles

- `owner`
- `worker`
- `moderator`
- `unverified user`

## Owner

Can:

- create PVZ
- view own PVZ list
- update own PVZ when no active vacancy exists
- create vacancy for approved own PVZ
- view responses for own vacancies
- accept response
- cancel assignment
- complete assignment
- read and write in own conversations while chat is writable
- create review
- create review reply

Cannot:

- create vacancy for non-approved PVZ
- update PVZ when active vacancy exists
- write in finalized chat

## Worker

Can:

- browse vacancies
- create response
- withdraw response before assignment
- read and write in own conversations while chat is writable
- cancel assignment
- create review
- create review reply

Cannot:

- respond twice to same vacancy
- withdraw response after assignment exists
- write in finalized chat

## Moderator

Can:

- approve PVZ
- reject PVZ
- block PVZ
- inspect moderation queue

Cannot:

- create assignments
- write reviews as part of marketplace business flow

## Unverified user

Has valid JWT but restricted access.

Allowed:

- `POST /v1/auth/logout`
- `POST /v1/auth/resend-code`
- `POST /v1/auth/verify-email`
- `GET /v1/me` is forbidden until email verification succeeds

Denied:

- PVZ actions
- vacancy actions
- response actions
- chat
- reviews

## Review permissions

- review can be created only by assignment participant
- `target_user_id` must be the second assignment participant
- reply can be created only by assignment participant
- reply is attached only to root review

## Chat permissions

- only owner and worker from conversation can read it
- only owner and worker can write while chat is writable
- chat is read-only after assignment finalization
