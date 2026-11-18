# events

The event log of the core. It is also the outbox: an event and its River job commit in the same transaction as the action ([ADR 0003](../../../docs/dev/adr/0003-analytics-outbox-and-sinks.md), [ADR 0014](../../../docs/dev/adr/0014-river-and-outbox.md)).

Tables: events

## Boundaries

- Now: `Recorder.Record` writes an event and a `events.dispatch` job in the caller's transaction. `DispatchWorker` reads the event in the account context and gives it to each `Sink`.
- No sink exists yet: the analytics tool is not selected ([questions](../../../docs/dev/questions.md#infrastructure-choices)). Reports for masters read the table directly, through their own module interface.
- The list of event names is [docs/dev/events.md](../../../docs/dev/events.md). The table checks the name form.

## Interface

- `NewRecorder(riverClient)`, `Recorder.Record(ctx, tx, accountID, Event)`. Call it inside `db.WithAccount`, with the transaction of the action.
- `NewDispatchWorker(pool, logger, sinks...)`: the worker process registers it.
- `Sink`: the port for an external tool. A sink filters what may leave the core and accepts the same event twice.

## Delivery

- No event is lost: the event row and the job are in one transaction.
- A failed sink fails the job, and River retries it with a back-off. After the last attempt the job is "discarded". The dead-letter alert of backend standard 06 watches this.
- No sweep job and no `dispatched_at` column: River already gives the guarantee (ADR 0014).

## Dependencies

`platform/db`, River.
