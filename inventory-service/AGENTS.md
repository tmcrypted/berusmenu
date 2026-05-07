# AGENTS

See the global guide first: [../AGENTS.md](../AGENTS.md)

## Service Role

`inventory-service` owns PVZ and moderation workflow.

## Local Rules

- Only this service mutates `inventory.*`.
- No vacancy creation rules are enforced by direct SQL from other services.
- Moderation events are append-only business history.
