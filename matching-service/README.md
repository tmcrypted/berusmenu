# matching-service

`matching-service` is the planned owner/worker marketplace service.

It is not implemented as a runnable backend service yet. Its protobuf contracts and architecture rules already exist so the inventory/auth work can integrate against stable boundaries.

## Responsibility

- vacancy lifecycle;
- response lifecycle;
- assignment lifecycle;
- reviews;
- review replies;
- active vacancy truth for PVZ update checks;
- response chat writability checks.

## Planned Owned Schema

Schema: `matching`

Tables:

- `matching.vacancies`
- `matching.responses`
- `matching.assignments`
- `matching.reviews`
- `matching.review_replies`

## Rules

- Vacancy can be created only for approved owner PVZ.
- PVZ approval is checked through `inventory-service`, not by reading `inventory.*`.
- Active vacancy for inventory update guard means `published` or `assigned`.
- Repeat response to the same vacancy after `withdrawn` is forbidden.
- Assignment is created when owner accepts a response.
- Old vacancy is not reused after `assignment.cancelled`.
- Reviews are allowed after `completed` and `cancelled`.
- Review replies are flat: only one reply level under root review.
- `matching.response_created` is published after response creation.

## Required Inventory Integration

`matching-service` must implement:

```text
HasActiveVacancyForPvz(pvz_id)
```

Expected behavior:

- return `true` if at least one vacancy for the PVZ has status `published` or `assigned`;
- return `false` for `completed`, `cancelled`, or no vacancies;
- inventory uses this check before owner PVZ update and fails closed if the RPC is unavailable.

## Required Communication Integration

`matching-service` must expose:

```text
CheckResponseChatWritable(response_id)
```

`communication-service` will use it before accepting message writes.

## Next Steps

1. Add module/runtime structure following the existing service pattern.
2. Add migrations.
3. Implement vacancy CRUD/listing.
4. Implement responses and assignment creation.
5. Publish `matching.response_created`.
6. Implement active vacancy checker for inventory.
7. Implement chat writability checker for communication.
8. Implement reviews and replies.
