# AGENTS

See the global guide first: [../AGENTS.md](C:/Users/tmcrypted/Desktop/berusmenu/AGENTS.md)

## Repo Role

`communication-service` owns chat state and message history.

## Local Rules

- No business assignment logic here.
- Chat write permissions depend on assignment state provided by owning workflow.
- `last_message_at` should reflect latest chat activity.
