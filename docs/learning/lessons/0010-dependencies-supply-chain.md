# 0010 — Code that we did not write runs in our product: the supply chain

From PRs #55, #68, #88, #89, #96 · 12.11.2026 · Topics: security, dependencies, CI

## Problem

Most of the code that runs in the product and in CI is not ours: Go modules, npm packages, GitHub Actions. A problem in one of them becomes our problem. Our own repository can also leak a secret, and then others can use it.

## Predict first

Nobody changed the repository for a week. On Monday, the dependency audit fails on each PR. How is that possible?

## Theory

The **supply chain** is all the code that reaches your product from other people. It has more parts than `package.json` shows:

- **Direct dependencies** are the packages that your code imports.
- **Transitive dependencies** are the packages that your dependencies import. In Go, `go.mod` marks them `// indirect`. In npm, they are deep in the lockfile.
- **Build tools and CI actions** run with access to your code and your tokens.

Three things go wrong:

- **A known vulnerability.** Somebody finds a bug in a version that you use and publishes an advisory. Your code did not change, but the list of known bugs did. An audit tool compares your versions with that list.
- **A malicious release.** An attacker publishes a bad version, or moves a tag to other code. A version range (`^7`) or a tag (`@v7`) takes the new code automatically. A lockfile hash or a commit SHA cannot change. A **cooldown** helps too: wait some days before you take a new release, because a bad release is usually found and removed in that time.
- **A leaked secret.** A token in a commit stays in the git history after you delete the file. A secret scanner must read all commits.

Sometimes no fixed version exists for a deep dependency. Then you have two tools. An **override** forces a safe version for the whole tree. An **accepted finding** records that you know the risk, with a reason and an end date.

Sources:

- [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use), GitHub Docs: why a full commit SHA is the only fixed reference to an action.
- [We should all be using dependency cooldowns](https://blog.yossarian.net/2025/11/21/We-should-all-be-using-dependency-cooldowns), William Woodruff: the case for a cooldown.
- [Vulnerability Management for Go](https://go.dev/blog/vuln): how govulncheck reports only the vulnerable code that you call.
- [Dependabot options reference](https://docs.github.com/en/code-security/reference/supply-chain-security/dependabot-options-reference), GitHub Docs: the `cooldown` option.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Update when somebody remembers | No tools | Known vulnerabilities stay for months |
| An audit on each PR and on a schedule | A tool compares the versions with the advisory lists | It needs the network. A new advisory makes an unrelated PR red |
| Pin CI actions by commit SHA | The reference cannot move | A SHA is not readable, so a comment gives the version. A bot must update it |
| A cooldown before each update | The update bot waits 7 days after a release | A security fix also waits, unless you update it by hand |
| An override for a deep dependency | The package manager forces one version for the whole tree | The parent package can break with a version that it does not expect |
| An accepted finding with a reason and an end date | The audit ignores one ID until the date | Somebody must look at it again when the date comes |

## Edge cases

| What if | What happens |
| --- | --- |
| The newest version of a tool still pins the vulnerable module | An update of the tool does not help. We raised the indirect module with its own `go get` |
| An override breaks the tool | A new faker breaks the Prism mock server. We did not override it. We accepted the finding with a reason and the end date 04.05.2027 |
| The end date passes | The audit fails again, so somebody looks at the finding again |
| A new advisory appears and the repository did not change | The audit on each PR fails. A weekly audit finds it also when nobody opens a PR |
| A security release is one day old | The cooldown is for the bot. We took Go 1.27.4 immediately, because it was the official security release |
| The vulnerable function is in a module that we use, but our code does not call it | govulncheck reads the call graph and does not report it. A lockfile scanner reports it, because it sees only versions |
| A secret is in an old commit and the file is deleted | The secret is still in the history. The scan reads all commits, and the commit hook reads the files before they get there |

## Decision

We use all of these together, each for its own threat. govulncheck audits the Go code, and osv-scanner audits the npm lockfiles, on each PR and each week. The audits need the network, so they are not part of the offline lint. Each GitHub Action is pinned by a commit SHA, and a linter fails on a tag. Dependabot waits 7 days after a release. Overrides fix deep dependencies of dev tools, with a comment that says when to remove each entry. gitleaks scans the full history in CI and the changed files before a commit.

An accepted finding is the last option. Ours is safe only because the package is in a dev tool and the vulnerable function is not called. It would be wrong for a package that the product runs.

Record: `docs/dev/backend-standard/05-security.md` and the gates table of [lesson 0009](0009-fast-checks-for-agents.md).

## In our code

1. `.github/dependabot.yml`: the cooldown, and the updates that stay manual.
2. `.github/workflows/ci.yml`, a `uses:` line: a commit SHA with the version in a comment.
3. `web/pnpm-workspace.yaml`, `overrides`: the forced versions and the rule to remove them.
4. `osv-scanner.toml`: the accepted finding, its reason and `ignoreUntil`.

## Check yourself

1. Why is `actions/checkout@v7` weaker than a commit SHA, when both name "version 7"?
2. Why does an accepted finding need an end date, and not only a reason?
3. An update of a tool to its newest version did not close the alert. Where do you look?

<details><summary>Answers</summary>

1. A tag is a name that the owner of the action can move to other code, also after an attack on their account. Your workflow then runs the new code with your tokens. A commit SHA names one exact state of the code and cannot move.
2. The reason is true only for now. The code can start to call the vulnerable function, or a fixed version can appear. The end date makes the check fail again, so a person must decide again.
3. In the transitive dependencies. The newest tool can still pin the old vulnerable module. In Go you raise that module directly with `go get`. In pnpm you add an override, and you check that the tool still works.

</details>
