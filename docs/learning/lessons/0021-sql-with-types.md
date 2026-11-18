# 0021 — How the code talks to the database: ORM, query builder or SQL with generated types

From PR #106 · 12.11.2026 · Topics: database access, code generation

## Problem

The first module needs to read and write rows. Go has static types, and SQL is text. Something must connect the two: put Go values into a query and put the result columns into Go fields. The choice decides who writes the SQL, and when a mistake in it shows.

## Predict first

A migration renames a column. Ten queries in the code still use the old name. With SQL as strings in the code, when do you learn about it? With code that a tool generates from the schema and the queries, when do you learn about it?

## Theory

There are three common levels.

- **ORM.** You work with objects, and the library writes the SQL. Django ORM and Prisma are examples. Simple operations are short. The cost shows later: you do not see the real query, a loop can send one query for each row, and a database feature that the ORM does not know needs a way around it.
- **Query builder.** You build SQL from function calls, as with Knex or Kysely. You control the query, and you can add parts by condition. You still write in a second language that copies SQL.
- **Plain SQL.** You write the query as SQL. By hand, you also write the code that copies each column into a field. A wrong name or a wrong order shows only when the code runs.

**Code generation from SQL** closes the gap of the third level. A tool reads the schema and each named query. It checks the query against the schema and writes one typed function for it: the parameters as arguments, the row as a struct. It is the "contract first" idea for the database. `openapi-typescript` makes types from an API description in the same way. A wrong column name is then an error at generation time, not in production.

The weak point of this level is a query whose shape changes at run time, for example a list with ten optional filters. Each shape needs its own written query.

Sources:

- [sqlc documentation](https://docs.sqlc.dev/en/latest/): how the tool compiles SQL into typed code.
- [Getting started with PostgreSQL](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html), the sqlc tutorial: one query from SQL to a Go function.
- [Accessing relational databases](https://go.dev/doc/database/), the Go documentation: the levels of database access in Go.
- [How We Went All In on sqlc/pgx for Postgres + Go](https://brandur.org/sqlc), Brandur Leach: a comparison with the other Go options from real use.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| ORM (GORM, ent) | Structs and methods. The library writes the SQL | Hidden queries. Our row rules and composite keys need full control of the SQL |
| Query builder | Go calls build the SQL text | A second language. The types of the result are still your work |
| SQL by hand with pgx | SQL strings and manual scans | A mistake shows only at run time. Much repeated code |
| SQL by hand, sqlc generates the Go code | The query is in a `.sql` file. The tool writes a typed function | A generation step. Dynamic queries are hard |

## Edge cases

| What if | What happens |
| --- | --- |
| Somebody changes a query and does not generate again | CI runs the generation and fails when the files in the repository differ |
| Somebody edits a generated file | The next generation removes the edit. The source is the `.sql` file |
| The tool does not know a Go type for a column | A type override in the config maps it. We map `uuid` to the `uuid` type of the Go standard library |
| A query that must return one row finds none | The function returns the error `pgx.ErrNoRows`. The code changes it into its own error |
| The schema changes | The tool reads the schema from the migration files, so no second copy of the schema exists |
| A query must run inside a transaction | The generated code accepts a pool or a transaction. The module passes the transaction that carries the account |
| Another module wants to use the queries | It cannot. Each module has its own `queries/` folder and a private `store` package |

## Decision

We write SQL by hand, and sqlc compiles it into typed Go code on top of pgx. We do not use an ORM. The SQL in the repository is the SQL that runs, so a review can read it and `EXPLAIN` can check it. The account filter and the row rules are visible in each query.

This would be wrong for an application that is mostly generic forms over many tables, such as an admin panel. There an ORM saves much code. It is also a poor fit for search screens with many optional filters.

Record: [backend standard 03](../../dev/backend-standard/03-data.md), section "Queries". The type override came from the stress round of [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `core/internal/accounts/queries/accounts.sql`: the queries, each with a name and a result kind.
2. `core/internal/accounts/store/accounts.sql.go`: the generated functions. Compare one with its query.
3. `core/sqlc.yaml`: the schema source, one entry for each module, and the type overrides.
4. `core/internal/accounts/accounts.go`: how the service calls `store.New(tx)` inside `db.WithAccount`.

## Check yourself

1. A migration renames `expires_at`. With sqlc, where does the mistake in an old query show?
2. Why does CI generate the code again, when the generated files are already in the repository?
3. You must build a client list with eight optional filters. Why is this hard with our choice, and what are the ways out?

<details><summary>Answers</summary>

1. At generation time. The tool reads the new schema from the migrations and fails on the query that uses the old name. If the result type changes, the Go compiler then fails on the code that uses it. Nothing waits until the query runs.
2. A person can change a `.sql` file or a migration and forget to generate. Then the Go code and the SQL do not match. CI generates and compares, so a stale file stops the merge.
3. sqlc makes one function for one fixed query. Eight optional filters give many shapes. The ways out: one query in which each filter has the form "the parameter is null or the column matches", or a small query builder for that one screen.

</details>
