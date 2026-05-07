# AGENTS

See the global guide first: [../AGENTS.md](../AGENTS.md)

## Service Role

`notification-service` is an async delivery worker.

## Local Rules

- No domain ownership beyond notification delivery.
- Do not move business validation here.
- Consumers react to events that already happened.
- Email channel only in MVP, but internal sender abstraction should stay channel-agnostic.
