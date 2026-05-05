# communication-service

Сервис чатов и переписки по response flow.

## Ответственность

- create conversation for response
- store messages
- track message reads
- support realtime delivery semantics

## Owned schema

Schema: `communication`

Tables:

- `communication.conversations`
- `communication.messages`
- `communication.message_reads`

## Важные правила

- conversation создается сразу после response
- conversation is unique per response
- only owner and worker of conversation can read and write
- chat becomes read-only after assignment finalization

## Integrations

- PostgreSQL
- Redis for presence/typing/active connection state
- websocket traffic through gateway

## Next implementation steps

1. conversation creation
2. message persistence
3. read markers
4. websocket delivery
