# 0022 — Which value identifies a row: serial numbers, random UUIDs or time-ordered UUIDs

From PR #106 · 12.11.2026 · Topics: data modeling, indexes, identifiers

## Problem

Each row needs a primary key. The key appears in URLs, in logs and in other tables. The first tables of the project, `accounts` and `events`, had to select a kind of key, and each later table copies that selection. A change later means a rewrite of each foreign key.

## Predict first

A public URL contains `/clients/1042`. What does a competitor learn from this number, and what can they try next? Now the ID is a random UUID. What do they learn? What does the database pay for the random value on each insert?

## Theory

There are three common kinds of key.

- **Serial number** (`bigserial`, auto-increment). The database gives 1, 2, 3. It is small and ordered. But it shows how many rows exist and how fast the number grows. The next ID is easy to guess. Only the database can make it, so the code learns the ID after the insert.
- **Random UUID (version 4).** 128 bits, of which 122 are random. Each process can make one without the database, and nobody can guess the next one. It shows nothing about the data.
- **Time-ordered UUID (version 7).** The first 48 bits are the time in milliseconds, and the other bits are random. New IDs sort after old ones. Each process can make one. It shows the creation time of the row, but not the number of rows.

The order matters because of the index. A primary key has a **B-tree** index: a sorted tree of pages. A key that always grows goes to the last page, so the database works with a few pages that stay in memory. A random key goes to a random page. On a large table each insert touches a different page, the pages split and stay half empty, and the index becomes larger and slower. Version 7 keeps the good property of a serial number for the index, and the "make it anywhere" property of a UUID.

No kind of ID is a permission. A hard-to-guess ID reduces what an attacker can list, but the server must still check that the row belongs to the caller.

Sources:

- [RFC 9562: UUIDs](https://www.rfc-editor.org/rfc/rfc9562): versions 4 and 7, and the section "Security Considerations".
- [PostgreSQL: UUID Type](https://www.postgresql.org/docs/16/datatype-uuid.html): the `uuid` column type.
- Kleppmann, "Designing Data-Intensive Applications", chapter 3, section "B-Trees": pages, splits and why the write position matters.
- [Goodbye to sequential integers, hello UUIDv7!](https://buildkite.com/resources/blog/goodbye-integers-hello-uuids/), Buildkite: a move between the three kinds on a large Postgres database.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Serial number | The database counts | It shows the number of rows. IDs are easy to guess. The code gets the ID after the insert |
| UUID version 4 | 122 random bits | Random inserts into the index. No order by time |
| UUID version 7 from the database (`uuidv7()`) | A column default makes the ID | The function needs Postgres 18. We run Postgres 16 |
| UUID version 7 from google/uuid | The Go code makes the ID | A dependency with no release since 2024 |
| UUID version 7 from the Go standard library | The Go code makes the ID | It needs Go 1.27, and the SQL generator needs a type override |

## Edge cases

| What if | What happens |
| --- | --- |
| The code needs the ID before the row exists | It has it. To create an account, the code makes the ID, sets it as the account of the transaction, and then inserts the row. The row rule accepts the insert only then |
| Somebody reads the time from an ID | They learn when the row was made, to the millisecond. If that is a secret, the row needs a different public ID |
| A value must be a secret | Do not use an ID. The session token is a separate random value, and its hash is the key of `sessions` ([lesson 0019](0019-stolen-database.md)) |
| The server clock goes back | The new IDs sort before some older ones. They are still unique. Do not use the ID as an exact order of events |
| A list needs a stable order for pages | The cursor uses `(created_at, id)`. The ID breaks the tie between rows with the same time |
| The SQL generator does not know the new Go type | The stress round of the ADR expected this. An override in `core/sqlc.yaml` maps the `uuid` column to the standard library type |

## Decision

IDs are UUID version 7, made in the Go code with the `uuid` package of the standard library. The column type is `uuid`, with no default in the database. The standard gives the reasons: the IDs are ordered by time, work well in indexes and do not show the number of records.

This would be wrong for a table where the creation time is private and the ID is public. It also costs 16 bytes for each key and each foreign key, against 8 for a `bigint`. For a small internal table with no public ID, a serial number is sufficient.

Record: [backend standard 03](../../dev/backend-standard/03-data.md), section "Schema". The library choice is in [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `core/internal/accounts/accounts.go`, `CreateAccount`: the ID exists before the insert, and the account context uses it.
2. `core/migrations/00002_accounts_sessions.sql`: `id uuid PRIMARY KEY` with no default, and the row rule that needs the ID.
3. `core/sqlc.yaml`: the override for the `uuid` type.
4. [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md): the row "UUIDv7" and the first line of the stress round.

## Check yourself

1. Why does a random UUID make inserts slower on a large table, and why does version 7 not have this problem?
2. What does a version 7 ID still show to a person who sees it?
3. An API returns a row to any caller who knows its UUID. Is this safe, because nobody can guess a UUID?

<details><summary>Answers</summary>

1. The index is a sorted tree. A random value goes to a random page, so each insert reads and changes a different page, and pages split in the middle. A version 7 ID starts with the time, so new values go to the end of the index, on pages that are already in memory.
2. The creation time of the row, in the first 48 bits. It does not show how many rows exist or what the next ID is.
3. No. An ID is a name, not a permission. IDs appear in URLs, logs and messages, so other people can see them. The server must check that the row belongs to the account of the caller. In our project the account rules of the database do this.

</details>
