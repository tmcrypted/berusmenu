# AGENTS

See the global guide first: [../AGENTS.md](../AGENTS.md)

## Service Role

`matching-service` owns the staffing workflow.

## Local Rules

- Only this service changes vacancy, response, assignment and review state.
- `vacancy_id unique` in assignment assumes cancelled vacancy is not reused.
- Review reply is always attached to root review, never to another reply.
