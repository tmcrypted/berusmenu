# AGENTS

See the global guide first: [../AGENTS.md](../AGENTS.md)

## Service Role

`api-gateway` is the public HTTP entrypoint for Berusmenu.

## Local Rules

- Keep gateway thin.
- Business rules stay in owning services.
- Identity must come from JWT and flow downstream through gRPC metadata.
- Unverified users must be blocked from non-whitelist endpoints in gateway middleware.
