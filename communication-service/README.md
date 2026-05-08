# communication-service

`communication-service` is the planned chat service for the response/assignment flow.

It is not implemented as a runnable backend service yet. Its boundaries and protobuf contracts already exist.

## Responsibility

- create conversation for response;
- store messages;
- track message reads;
- own chat permissions and read state;
- support realtime delivery semantics with gateway/websocket edge.

## Planned Owned Schema

Schema: `communication`

Tables:

- `communication.conversations`
- `communication.messages`
- `communication.message_reads`

## Rules

- Conversation is created immediately after response.
- Conversation is unique per response.
- Only owner and worker from the conversation can read/write.
- Chat becomes read-only after `assignment.completed` or `assignment.cancelled`.
- Message writes must check response chat writability through `matching-service`.

## Integrations

- PostgreSQL for conversations/messages/read state.
- Redis for presence, typing and active websocket state.
- `matching-service.CheckResponseChatWritable` for write permission.
- `api-gateway` websocket edge for public traffic.

## Next Steps

1. Add module/runtime structure following existing service pattern.
2. Add migrations.
3. Implement `CreateConversationForResponse`.
4. Implement message persistence and listing.
5. Implement read markers.
6. Implement realtime delivery path.
