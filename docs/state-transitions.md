# State Transitions

## `auth.users.status`

States:

- `active`
- `blocked`

Rules:

- Moderator spam action sets `status = blocked` and `blocked_until = now + N minutes`.
- Login, refresh and JWT validation are denied while `blocked_until` is in the future.
- After `blocked_until` expires, the user can authenticate again without a separate unblock action.
- Moderators cannot block themselves or other moderators.

## `inventory.pvz.status`

States:

- `pending`
- `approved`
- `rejected`
- `blocked`

Transitions:

- create -> `pending`;
- moderator approve -> `approved`;
- moderator reject -> `rejected`;
- moderator block -> `blocked`;
- owner update -> `pending`;
- moderator spam cleanup -> pending rows deleted, remaining owner rows -> `blocked`.

Rules:

- update is forbidden if active vacancy exists with status `published` or `assigned`;
- new vacancy creation is forbidden unless PVZ is `approved`;
- active vacancy check is performed through `matching-service`;
- update resets `rejection_reason`, `moderated_by`, `moderated_at`;
- update appends `submitted` moderation event;
- moderation decisions append `approved`, `rejected`, or `blocked` events.

## `matching.vacancies.status`

Planned states:

- `published`
- `assigned`
- `completed`
- `cancelled`

Planned transitions:

- create -> `published`;
- accept response -> `assigned`;
- complete assignment -> `completed`;
- cancel assignment -> `cancelled`.

Rules:

- after `cancelled`, this vacancy is not reused.

## `matching.responses.status`

Planned states:

- `pending`
- `accepted`
- `rejected`
- `withdrawn`

Planned transitions:

- create -> `pending`;
- owner accept -> `accepted`;
- owner rejects or another response accepted -> `rejected`;
- worker withdraw before assignment -> `withdrawn`.

Rules:

- repeat response to same vacancy is forbidden after withdrawal.

## `matching.assignments.status`

Planned states:

- `active`
- `completed`
- `cancelled`

Planned transitions:

- create -> `active`;
- owner completes -> `completed`;
- owner or worker cancels -> `cancelled`.

Rules:

- `cancelled_by_user_id` required logically on cancel;
- reviews are allowed after `completed` or `cancelled`.
- owner completes active assignment;
- owner or worker can cancel active assignment;
- moderator can perform business-valid dispute actions through audited moderation routes.

## Chat State Rule

`communication.conversations` does not need an explicit status column in MVP.

Logical rule:

- writable while assignment is active or not yet created;
- read-only after `assignment.completed` or `assignment.cancelled`.
