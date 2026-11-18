# Learning from the project

13.11.2026 · @ilnar

Agents write the product code. The owner learns from the real problems that the work brings. Each problem that needed a choice becomes one lesson: the problem, the theory, the options, the edge cases and the decision.

A general course runs separately, outside this repository. These lessons hold only what the project really met.

## The lesson format

One lesson is one problem, not one PR. One PR can give several lessons. A PR with no real choice gives none. A lesson takes about five minutes to read. Each lesson has the same eight sections, in this order:

| # | Section | What it holds |
| --- | --- | --- |
| 1 | Problem | The situation in the product that needed a decision, in plain words |
| 2 | Predict first | One question. Answer it before you read on |
| 3 | Theory | The general model, not specific to this project. It ends with a "Sources" list: a book chapter, an article or official documentation |
| 4 | Options | A table: each option, how it works, its cost |
| 5 | Edge cases | A table: "what if" and what happens |
| 6 | Decision | What we selected, why, and when the selection would be wrong. A link to the ADR, if one exists |
| 7 | In our code | Three or four files to read, in order |
| 8 | Check yourself | Two or three questions. The answers are in a collapsed block |

The first line under the title gives the PR, the date and the topics: `From PR #107 · 12.11.2026 · Topics: data consistency, messaging`.

The reference example is [lesson 0023](lessons/0023-dual-write-outbox.md).

## Study a lesson with an agent

Start an agent session in the repository and run `/lesson 0023`, or `/lesson` to select from the list. The agent is a tutor for 20 to 30 minutes:

1. It asks the "Predict first" question and waits for your answer.
2. It asks how you would solve the problem, before it shows the options.
3. It teaches the theory in small parts, with one question after each part.
4. It goes through the code and the edge cases. You answer "what happens if" before you see the answer.
5. You select an option and compare it with the decision of the project.
6. It asks the "Check yourself" questions and one question about a different business.

At the end you get a short result: what is solid, what is weak, and what to read. The agent does not change a file. Do the last step again after one week in a new session.

Commands in the session: `simpler`, `deeper`, `skip`, `stop`. The prompt of the tutor is the skill `.claude/skills/lesson/SKILL.md`.

## Rules for the agent that writes a lesson

- Write the lesson in the same PR as the change. The file is `lessons/NNNN-<topic>.md`, with the next free number. Add a row to the table below.
- Write a lesson only for a problem that had a real choice: more than one workable option, or an edge case that changed the solution.
- The title names the problem, not the tool: "Two systems must change together", not "River".
- The theory must stand without the project. A reader who never saw this code must understand it.
- Give two to four sources. Each source must exist: open each link before you add it. For a book, give the chapter.
- The reader knows the frontend and TypeScript well, wrote some Django and Node.js, and is new to Go. Use an analogy from those where it is honest.
- A "Predict first" question and a "Check yourself" question ask "what happens if" or "why". They do not ask for a name from the code.
- If a later change replaces a decision, change the lesson in the same PR.

## Lessons

The PR numbers are from the project that this template comes from. They do not exist in your repository.

| # | Lesson | Topics | PR |
| --- | --- | --- | --- |
| 0001 | [Where the backend, the frontends and the contract live: one repository or several](lessons/0001-monorepo-layout.md) | repository layout, tooling | #32 |
| 0002 | [Module boundaries decay when nothing checks them: automatic architecture checks](lessons/0002-module-boundaries.md) | architecture, modularity, linters | #40, #42 |
| 0003 | [Who changes the database schema and when: migrations as a separate step](lessons/0003-migrations-separate-step.md) | database migrations, deploy | #44 |
| 0004 | [One build for every environment: settings from the environment, checked at startup](lessons/0004-config-from-environment.md) | configuration, deploy, fail fast | #44, #105 |
| 0005 | [The application must not be able to do everything in the database: roles and least privilege](lessons/0005-database-roles.md) | database security, least privilege | #44, #106 |
| 0006 | [One database, many accounts: how account A never sees the data of account B](lessons/0006-tenant-isolation.md) | multi-tenancy, row-level security | #30, #46 |
| 0007 | [What a test must use to be trusted: mocks or a real database](lessons/0007-tests-with-real-database.md) | testing, integration tests, test isolation | #46 |
| 0008 | [How the backend and the frontend stay in agreement about the API: one contract, generated code](lessons/0008-contract-first-codegen.md) | API contract, code generation | #47 |
| 0009 | [How to trust code that an agent wrote: fast checks first](lessons/0009-fast-checks-for-agents.md) | quality gates, agent workflow | #55, #57, #61, #89 |
| 0010 | [Code that we did not write runs in our product: the supply chain](lessons/0010-dependencies-supply-chain.md) | security, dependencies, CI | #55, #68, #88, #89, #96 |
| 0011 | [What the API tells a client when something fails: one error format](lessons/0011-api-error-format.md) | API design, security | #62 |
| 0012 | [Can we go back: rollbacks and backups that somebody really tried](lessons/0012-rollback-and-backups.md) | migrations, operations | #65 |
| 0013 | [A schema change that waits for a lock stops the whole application](lessons/0013-migration-lock-timeout.md) | Postgres, migrations, locks | #89 |
| 0014 | [An old client must keep working after a deploy](lessons/0014-breaking-api-changes.md) | API design, compatibility, quality gates | #89 |
| 0015 | [Every request needs the same things: the layers around a handler](lessons/0015-layers-around-a-request.md) | HTTP middleware, error handling | #105 |
| 0016 | [Alive or ready: when to restart a process and when to stop its traffic](lessons/0016-alive-or-ready.md) | health checks, operations | #105 |
| 0017 | [A user reports an error: finding one request in logs, traces and error reports](lessons/0017-finding-one-request.md) | observability, logging, tracing | #105 |
| 0018 | [Who sends this request: server-side sessions or self-contained tokens](lessons/0018-sessions-or-tokens.md) | authentication, sessions | #106 |
| 0019 | [A stolen copy of the database must not give working logins](lessons/0019-stolen-database.md) | security, sessions, hashing | #106 |
| 0020 | [Another site makes the browser send a request with our cookie: CSRF](lessons/0020-requests-from-another-site.md) | security, CSRF, cookies | #106 |
| 0021 | [How the code talks to the database: ORM, query builder or SQL with generated types](lessons/0021-sql-with-types.md) | database access, code generation | #106 |
| 0022 | [Which value identifies a row: serial numbers, random UUIDs or time-ordered UUIDs](lessons/0022-primary-keys.md) | data modeling, indexes, identifiers | #106 |
| 0023 | [Two systems must change together: the dual write and the outbox](lessons/0023-dual-write-outbox.md) | data consistency, messaging | #107 |
| 0024 | [A job or a message can arrive twice: delivery guarantees and the idempotent receiver](lessons/0024-at-least-once-idempotency.md) | messaging, reliability | #107 |
| 0025 | [Work that the user must not wait for and that a restart must not lose: a job queue in the database](lessons/0025-queue-in-postgres.md) | background jobs, infrastructure choice | #107 |
| 0026 | [A long list goes to the client in pages: offset, keyset and the opaque cursor](lessons/0026-pages-of-a-long-list.md) | API design, SQL | #108 |
| 0027 | [Background work is stuck and no user sees an error: how to know that a queue is healthy](lessons/0027-is-the-queue-healthy.md) | observability, background jobs | #114 |
| 0028 | [The server returns an error: what the user sees](lessons/0028-server-errors-on-screen.md) | error handling, forms, routing | #109 |
| 0029 | [The same validation rule in the server and in the form](lessons/0029-validation-from-contract.md) | validation, contract, code generation | #109 |
| 0030 | [Two applications must look like one product](lessons/0030-design-tokens.md) | design tokens, theming, accessibility | #110, #117 |
| 0031 | [Interface texts in more than one language](lessons/0031-interface-texts.md) | i18n, bundle size, type safety | #111 |
| 0032 | [The first screen must load fast as the application grows](lessons/0032-routes-and-lazy-pages.md) | routing, code splitting, performance | #112 |
| 0033 | [People and agents must read the same documents](lessons/0033-docs-viewer.md) | docs as code, tooling choice, link checks | #127 |
| 0034 | [A sketch that people can change by hand](lessons/0034-editable-sketches.md) | diagrams as code, tooling choice, local write endpoints | #139 |
