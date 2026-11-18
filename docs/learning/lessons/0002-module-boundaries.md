# 0002 — Module boundaries decay when nothing checks them: automatic architecture checks

From PRs #40, #42 · 28.10.2026 · Topics: architecture, modularity, linters

## Problem

The backend is one program that we divide into modules. Each module has a public entry and its own tables. Agents write most of the code, and an agent takes the shortest way: it imports a private package of another module, or it reads the table of another module with a `JOIN`. Each such shortcut works today. After a hundred shortcuts, no module can change alone.

## Predict first

Module `clients` needs the list of services. The table `services` belongs to module `catalog`. An agent writes `SELECT ... FROM services` in the SQL of `clients`. The code compiles and the tests pass. What breaks later, and who finds out?

## Theory

A **module** is useful when its interface is much smaller than its implementation. Other code depends only on the interface, so the implementation can change freely. This is **information hiding**.

**Coupling** is how much one part knows about another. The weakest level is a contract: the caller knows only a published interface. The strongest level is intrusive: the caller knows private details, for example the tables. A database table is a private detail. If two modules read one table, a column change needs both authors, and nothing in the code shows that link.

A rule in a document does not hold a boundary. A boundary holds when a machine checks it on each change. Such a check is an **architecture test**, also named a fitness function: a test for the structure of the code, not for its behavior. The frontend analogy is an ESLint rule that forbids an import from one folder into another.

The input for these checks is cheap. The **import graph** says which package imports which. A SQL parser says which tables a query uses. A check must also fail at least one time on purpose. A check that never failed can be a check that does not run.

Sources:

- Ousterhout, "A Philosophy of Software Design", chapters 4 "Modules Should Be Deep" and 5 "Information Hiding (and Leakage)".
- Khononov, "Balancing Coupling in Software Design", and its summary at [coupling.dev](https://coupling.dev/): the strength, distance and volatility model.
- Ford, Parsons, Kua, "Building Evolutionary Architectures", chapter 2 "Fitness Functions": architecture rules as automatic tests.
- [Go 1.4 release notes, "Internal packages"](https://go.dev/doc/go1.4): the import guard that the compiler gives.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Rules in a document and in the review | The reviewer finds bad imports | The reviewer misses some. An agent trusts a rule that nothing checks |
| A ready tool (go-arch-lint) | A config file describes the permitted imports | A new dependency. It cannot check table ownership or count the coupling report |
| Our own small checker | About 100 lines of Go read `go list -json` and apply two import rules | We own the code and its tests |
| Table ownership with a regular expression | A regex finds table names in the SQL files | It read `SET` in an upsert as a table, and it missed quoted names |
| Table ownership with the Postgres parser | pg_query, the parser that sqlc uses, gives the real tables of each statement | It needs cgo. The first build takes about 30 seconds |

## Edge cases

| What if | What happens |
| --- | --- |
| The bad import is only in a `_test.go` file | The import rules apply to test imports too. The coupling report does not count them |
| A query uses a CTE (`WITH x AS ...`) and then reads from `x` | `x` is not a table. The checker tracks the CTE names for each statement, also in `WITH RECURSIVE` |
| A module has many more incoming links than the others | The coupling report prints it as a hotspot. The report does not block the merge |
| A linter can only give a warning | We do not use it. Each check blocks the merge or is silent, because people learn to ignore warnings |
| A check never failed | We do not trust it. We added a violation to the real tree, saw the failure, and removed the violation |
| A frontend package imports an app | oxlint fails: `web/packages/**` must not import from an app |

## Decision

We wrote our own checker, `archcheck`, and it runs in `mise run lint`, the same command that CI runs. It has two blocking checks and one report:

- **Import rules.** The core does not import a product module. No package imports the subpackage of another module.
- **Table ownership.** The README of each module has a line `Tables: a, b`. The SQL of a module can use only these tables.
- **Coupling report.** The mean number of links for each module, and the hotspots. The target for the mean is 1 to 3.

We selected our own tool because the rules are short and one tool does all three jobs. This would be wrong if the rules became many and complex. Then a ready tool with a rule language is cheaper than our code.

The human part stays: a coupling review for each new module, with the three axes from the theory. There is no ADR. The rules are in backend standard 01.

## In our code

1. `docs/dev/backend-standard/01-modules.md`: the module rules, the coupling model and the table of checks.
2. `core/tools/archcheck/rules.go`, `CheckImports` and `CouplingReport`: the import rules and the report.
3. `core/tools/archcheck/tables.go`, `UsedTables` and `CheckTables`: the tables of a query, compared with the README.
4. `core/internal/accounts/README.md`: a real `Tables:` line.

## Check yourself

1. Why is a `SELECT` on the table of another module a problem, when it changes no data?
2. Why does the import rule also apply to test files?
3. The coupling report does not block the merge, but the import rules do. Why are they different?

<details><summary>Answers</summary>

1. The query depends on the column names and the meaning of the rows. The owner module can no longer change its table without a change in the other module, and nothing in the Go code shows that link. The correct way is to call the public function of the owner.
2. A test that imports a private package of another module is the same link. A change in that private package then breaks the test of a different module. The checker had this gap, and the review found it.
3. An import rule has a clear answer: the import is permitted or it is not. The report is a number that needs judgment. A module with many links can be correct, for example a stable shared module. A check that blocks on judgment gives false failures.

</details>
