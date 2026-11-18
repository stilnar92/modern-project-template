# 0007 — What a test must use to be trusted: mocks or a real database

From PR #46 · 28.10.2026 · Topics: testing, integration tests, test isolation

## Problem

Most of our backend code is SQL, transactions and database rules. The most important promise of the product is that one account never sees the data of another ([lesson 0006](0006-tenant-isolation.md)). We had to decide what a test uses in place of the database, because that decides what a green test proves.

## Predict first

A test replaces the database with a mock that returns two rows. The real query has a typing error in a column name, and the real table has a rule that hides both rows. Does the test pass? What did it prove?

## Theory

A **test double** stands in for a real part. A **mock** is a double that returns the answers that the test author wrote. A test with a mock checks the code against the idea that the author has of the other part. If the idea is wrong, the test is green and the code is broken.

For a database, the idea is usually wrong in the places that matter. SQL syntax, constraints, transaction behavior and access rules exist only in the real engine. So an **integration test** uses our code and a real database together. Django does the same: its `TestCase` uses a real test database, not a fake ORM.

The old argument against this was the cost: somebody must install and clean a database. Containers removed most of it. A library starts the database in Docker from the test and removes it after. The test database can be the same engine, the same version and the same schema as production.

A real database is shared state, so tests must stay **isolated** from each other. The usual ways are a new database for each group of tests, a transaction that rolls back after each test, or data that cannot collide, for example new keys in each test.

Sources:

- Khorikov, "Unit Testing Principles, Practices, and Patterns", chapter 10 "Testing the database": why to use the real database, and how to keep the tests isolated.
- [Martin Fowler, "Mocks Aren't Stubs"](https://martinfowler.com/articles/mocksArentStubs.html): the kinds of test doubles and what each proves.
- [Testcontainers for Go, the Postgres module](https://golang.testcontainers.org/modules/postgres/): how a Go test starts Postgres in Docker.
- [Django documentation, "The test database"](https://docs.djangoproject.com/en/5.2/topics/testing/overview/): the same idea in a framework that you know.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Mock the database | The test gives fixed answers in place of queries | Fast, but it does not run the SQL, the transactions or the rules |
| A Postgres installed on each machine | Tests connect to a local server | Each developer and CI must install and clean it. Versions drift |
| One container for all test packages | One Postgres starts for the whole run | Packages share state and can break each other |
| One container for each test package | Each package starts a clean Postgres and Docker removes it after | About 2 seconds for each package. Tests need Docker |

## Edge cases

| What if | What happens |
| --- | --- |
| The test schema differs from the real schema | It cannot. The helper runs the same roles script as the local environment and applies all migrations |
| The test connects as a strong role | It proves nothing about the API. The helper gives three pools, and the isolation check uses the application role and the real entry point |
| An expected error stops the transaction | In Postgres, an error makes the rest of the transaction fail. The check runs the forbidden insert in a savepoint, a transaction inside a transaction |
| The forbidden insert fails for a different reason | A duplicate-key error is not proof of isolation. The check accepts only the row-level security error code. The review found this gap |
| Test setup must add rows without an account | The owner cannot: `FORCE` applies to it. Setup that must ignore the rules uses the system role |
| A check never failed | We do not trust it. Each check failed one time on purpose: a table with no rule, a table that leaks, a stopped API |

## Decision

We do not mock the database. The integration test with a real Postgres in a container is the main test level. Each test package gets its own container through one helper. The helper gives a migrated database and one connection pool for each role.

On top of that, two shared checks protect the isolation promise. One reads the Postgres system tables and fails when a table with `account_id` has no rule. The other plays account B against the rows of account A, for each table of each module.

Mocks stay correct for pure logic with no I/O, and for an external service that we cannot start in a container. For the frontend, a mock API made from the contract lets the work start before the backend is ready ([lesson 0008](0008-contract-first-codegen.md)).

There is no ADR. The rules are in backend standard 07.

## In our code

1. `docs/dev/backend-standard/07-testing.md`: the test levels and the rule "do not mock the database".
2. `core/internal/platform/testdb/testdb.go`, `Start`: the container, the roles script, the migrations and the three pools.
3. `core/internal/platform/testdb/isolation.go`, `CheckIsolation`: the four tries of account B.
4. `core/migrations/rls_test.go`, `TestRLS`: the check on the real migrated schema.

## Check yourself

1. Why does the isolation check use the application role and not the owner or the system role?
2. Why is "the insert failed" not enough to prove that the rule works?
3. Why did we make each check fail on purpose before we trusted it?

<details><summary>Answers</summary>

1. The real API works as the application role. The system role ignores the rules, so a test with it would show a leak that does not exist, or hide the fact that the rules work. The test must see what the API sees.
2. An insert can fail because of a duplicate key, a missing column or a broken constraint. Then the test is green for the wrong reason, and a table with no rule would pass. Only the specific row-level security error shows that the rule refused the row.
3. A check that always passes can be a check that does not run or that looks at the wrong tables. One failure on purpose proves that the check can see the problem that it exists for.

</details>
