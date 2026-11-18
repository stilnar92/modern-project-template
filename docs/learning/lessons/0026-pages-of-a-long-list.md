# 0026 — A long list goes to the client in pages: offset, keyset and the opaque cursor

From PR #108 · 12.11.2026 · Topics: API design, SQL

## Problem

A list endpoint can have thousands of rows: bookings, messages, events. The client must get them in pages. All list endpoints must do this in the same way, so we needed one rule and one helper before the first list endpoint exists.

## Predict first

A list shows the newest rows first, 20 on a page. The user reads page 1. Then one new row arrives. The user asks for page 2 with "skip 20 rows, give the next 20". Which row does the user see twice? And what happens if a row of page 1 is deleted and not added?

## Theory

**Offset pagination** says "skip N rows, give the next M": `LIMIT 20 OFFSET 40`. It is simple, and it lets a user go to page 50 directly. It has two problems:

- **It gets slow.** The database cannot go to row 10 000 directly. It reads and drops all the rows before it, so each next page costs more.
- **The pages move.** The offset counts from the top of the list. A new row at the top moves each row down by one, so the last row of page 1 comes again on page 2. A deleted row moves the list up, and the user misses a row.

**Keyset pagination** (also "seek" or "cursor" pagination) says "give the next M rows after this row". The client sends the sort key of the last row that it saw. The query is `WHERE (created_at, id) < (?, ?) ORDER BY created_at DESC, id DESC LIMIT 20`. An index on these columns finds the start point directly, so page 500 costs the same as page 1. New rows at the top do not move the start point. This is how the "load more" button of a feed works.

Keyset pagination needs a sort key that is **unique**. Two rows can have the same `created_at`. With only this column, the database can return them in any order, and a row can come twice or not at all. So the key gets a second column that breaks the tie: the `id`.

The **cursor** is the key in the form that the client gets. If the client can read it, the client starts to build it, and then the server cannot change its form. An **opaque** cursor is text with no meaning for the client, for example base64 of some bytes. The client only sends it back.

An opaque value needs a **canonical form**: one value has exactly one text. Many decoders accept more than one text for the same bytes. Then two different cursors mean the same position, and the server accepts text that it never made.

Sources:

- [Paging Through Results](https://use-the-index-luke.com/sql/partial-results/fetch-next-page), Markus Winand, "Use The Index, Luke": the offset method and the seek method, with the row comparison.
- [PostgreSQL: LIMIT and OFFSET](https://www.postgresql.org/docs/current/queries-limit.html): the skipped rows are still computed, and the order must be unique.
- [RFC 4648, section 3.5 "Canonical Encoding"](https://www.rfc-editor.org/rfc/rfc4648#section-3.5): why one base64 value can have more than one text. Section 3.3 is about characters that a decoder skips.
- [Go Fuzzing](https://go.dev/doc/security/fuzz/): the test type that found our bug.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| `OFFSET` and `LIMIT` | Skip N rows | Slow on far pages. A new or deleted row moves the pages |
| Keyset on `created_at` only | Start after the time of the last row | Rows with the same time come twice or are lost |
| Keyset on `(created_at, id)`, readable cursor | The client sends the time and the ID | Clients build cursors, so the form cannot change |
| Keyset on `(created_at, id)`, opaque cursor | The client sends back a text that only the server reads | No "go to page 50". A decoder to write and to test |

## Edge cases

| What if | What happens |
| --- | --- |
| A new row arrives between page 1 and page 2 | Page 2 starts after the last row of page 1, so it does not change. The new row is at the top, for the next reload |
| Many rows have the same `created_at` | The `id` breaks the tie. The test pages through 23 such rows, and each row comes one time |
| The server must know if a next page exists | The query asks for one row more than the limit. The extra row means "yes", and the server does not return it. No `count(*)` |
| It is the last page | `next_cursor` is empty |
| The client sends a cursor that the server did not make | The server answers 400. It does not guess |
| The cursor has a line break in the middle | The base64 decoder of Go skips line breaks, so this text decoded to a valid key. The decoder now encodes the key again and accepts the text only if it is the same |
| `limit` is missing, zero or 5000 | The default is 20, and the maximum is 100 |

## Decision

We use keyset pagination on `(created_at, id)`, newest first, with an opaque base64url cursor. It is our own small helper: ADR 0012 found no maintained library that fits our database tools. The cursor keeps the time in microseconds, which is the precision of the Postgres column, so a key goes to the database and back with no loss.

`Decode` accepts only the exact text that `Encode` makes. A fuzz test found the line-break problem in the first version in a few seconds. Go keeps the input that failed as a file, and runs it as a normal test from now on. A fuzz test is like a property-based test with fast-check: the tool makes the inputs.

This would be wrong for a screen that needs "page 37 of 120" or a jump to a page number. Keyset pagination cannot do that. Such a screen needs offset pagination on a small list, or a search and filters. Also, no list endpoint exists yet, so the helper is tested on a real table but not on a real endpoint.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), [backend standard 02](../../dev/backend-standard/02-api.md), section "Lists".

## In our code

1. `core/internal/platform/cursor/cursor.go`, the package comment: the SQL pattern and the index. Then `Encode`, `Decode` and `Page`.
2. `core/internal/platform/cursor/cursor_test.go`, `TestPagesThroughPostgres`: 23 rows in 3 pages. Then `FuzzDecode`.
3. `core/internal/platform/cursor/testdata/fuzz/FuzzDecode/`: the input that the fuzz test found.
4. `docs/dev/backend-standard/02-api.md`, section "Lists": the contract that the client sees.

## Check yourself

1. Why does a new row at the top of the list break offset pagination and not keyset pagination?
2. Why is `created_at` alone not a correct key, when the query has `ORDER BY created_at`?
3. Two different texts decode to the same cursor. No user sees a wrong page. Why is it still a defect?

<details><summary>Answers</summary>

1. An offset is a position from the top, and a new row changes each position. The last row of page 1 moves to position 21, which is on page 2, so the user sees it twice. A keyset cursor is a value, not a position: "after this row". Rows above it do not change what comes after it.
2. Two rows can have the same time. The database gives no fixed order for them. With "after this time", the second row of the pair is skipped if the page ended between them. With "at or after this time", the first row comes twice. A second, unique column makes each key different.
3. The server accepts text that it did not make, so "a cursor is valid only if we made it" is false. A cache or a log treats the two texts as two different requests. The strict rule "one key, one text" is easy to check, and it found a class of inputs that nobody planned for.

</details>
