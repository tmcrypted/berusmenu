# State Transitions

## `inventory.pvz.status`

States:

- `pending`
- `approved`
- `rejected`
- `blocked`

Transitions:

- create -> `pending`
- moderator approve -> `approved`
- moderator reject -> `rejected`
- moderator block -> `blocked`
- owner submits approved PVZ changes -> `pending`

Rules:

- submit update forbidden if active vacancy exists with status `published` or `assigned`
- new vacancy creation forbidden unless PVZ is `approved`

## `matching.vacancies.status`

States:

- `published`
- `assigned`
- `completed`
- `cancelled`

Transitions:

- create -> `published`
- accept response -> `assigned`
- complete assignment -> `completed`
- cancel assignment -> `cancelled`

Rules:

- after `cancelled` this vacancy is not reused

## `matching.responses.status`

States:

- `pending`
- `accepted`
- `rejected`
- `withdrawn`

Transitions:

- create -> `pending`
- owner accept -> `accepted`
- owner rejects or another response accepted -> `rejected`
- worker withdraw before assignment -> `withdrawn`

Rules:

- repeat response to same vacancy is forbidden

## `matching.assignments.status`

States:

- `active`
- `completed`
- `cancelled`

Transitions:

- create -> `active`
- owner completes -> `completed`
- owner or worker cancels -> `cancelled`

Rules:

- `cancelled_by_user_id` required logically on cancel
- reviews are allowed after `completed` or `cancelled`

## Chat state rule

`communication.conversations` does not have explicit state column in MVP.

Logical rule:

- writable while assignment is active or not yet created
- read-only after `assignment.completed` and `assignment.cancelled`
