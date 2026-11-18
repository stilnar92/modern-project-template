# Event dictionary

The list of the core events ([ADR 0003](adr/0003-analytics-outbox-and-sinks.md), [backend standard 04](backend-standard/04-async-integrations.md)). Only the orchestrator changes this file. A hook enforces this rule ([AI standard 02](ai-standard/02-harness.md#hooks)).

## Rules

- Record an event with `events.Recorder.Record(ctx, tx, accountID, event)` inside `db.WithAccount`, in the transaction of the action. The event and its dispatch job commit together ([ADR 0014](adr/0014-river-and-outbox.md)).
- Add an event here in the same PR as the code that writes it.
- The name has the form `<module>.<thing>_<past tense verb>`, for example `module.item_created`.
- Each event has these fields: `id`, `name`, `version`, `occurred_at`, `account_id`, and `data`.
- Only add fields to an event. A breaking change is a new version. Keep the old version while somebody reads it.
- The "Sink" column tells if the sink adapter sends the event to the analytics tool. Personal data of clients never goes to the sink.

## Events

No events exist yet.

| Name | Version | Module | Fields in `data` | Sink | Personal data |
| --- | --- | --- | --- | --- | --- |
