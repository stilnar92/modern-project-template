---
thesaurus-format: "2.0"
skill: ubiquitous-language
---

# Project Thesaurus

This file is the only glossary of the project. Code, the database, the API and the docs use these terms. `Example` lines give the word that the users of your product say. The template has only the terms of the foundation. Add your domain terms with the `ubiquitous-language` skill.

The `ubiquitous-language` skill maintains this file ([.claude/skills/ubiquitous-language](../../.claude/skills/ubiquitous-language/SKILL.md)). Search it before you name a thing: `rg -n 'Name' docs/dev/THESAURUS.md`. Only the orchestrator changes this file.

Sources: [CONTEXT.md](../../CONTEXT.md), [ADR 0003](adr/0003-analytics-outbox-and-sinks.md), [ADR 0005](adr/0005-api-surfaces-and-ownership.md).

## Index

- **Account** `Account` kind:entity avoid: `Tenant`
- **Background job** `BackgroundJob` kind:concept
- **Event** `Event` kind:entity
- **Session** `Session` kind:entity avoid: `Token`

## Terms

### Account

- **Definition**: The one who pays and owns the data. The tenant of the platform.
- **NOT**: A login identity (see `Owner` under Unresolved).
- **Related**: Session, Event

### Background job

- **Definition**: An execution mechanism for asynchronous work, in the River queue.
- **NOT**: A business concept with a cost or a result that a user sees.
- **Related**: Event

### Event

- **Definition**: A record in the analytics log. The core writes it in the same transaction as the action (outbox).
- **NOT**: A name for a domain event type. Each domain event gets its own Index line with `kind:event`.
- **Related**: Account, Background job

### Session

- **Definition**: A server-side record of one login to an account. The browser holds only an opaque cookie value.
- **NOT**: A self-contained token. The server can end a session at any time.
- **Related**: Account

## Forbidden

- `Tenant` use: `Account` — one name for the owner of the data

## Legacy

## Unresolved

### Owner — the login identity of an account

- **Found in**: [CONTEXT.md](../../CONTEXT.md) ("An account is a tenant, not a login identity")
- **Question**: Which term names the person who logs in and acts for an account?
- **Impact**: The login slice and the session model
- **Options**: `Owner`; `Member`; `User`
