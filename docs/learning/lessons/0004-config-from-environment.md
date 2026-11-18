# 0004 — One build for every environment: settings from the environment, checked at startup

From PRs #44, #105 · 12.11.2026 · Topics: configuration, deploy, fail fast

## Problem

The same program runs on a developer machine, in tests and in production. The database address, the passwords and the log level are different in each place. We must decide where these values live, and what the program does when one is missing or wrong.

## Predict first

The deploy forgets the variable `DATABASE_URL`. The code reads it only when the first request needs the database. When does somebody find the problem, and who is that person? What changes if the program reads all settings in its first second?

## Theory

**Configuration** is everything that changes between environments. Code is what stays the same. The twelve-factor rule is strict: you build one artifact, and you add the configuration when you run it. So the build that passed the tests is the same build that goes to production.

The usual carrier is **environment variables**. Each operating system and each container platform has them. They do not get into git by accident, as a config file can. You know them as `process.env` in Node.js and `os.environ` in Django settings.

Environment variables are strings with no schema. `PORT=abc` is a valid variable. The question is when the program finds out:

- **Read at use.** Each part of the code reads its variable when it needs it. A wrong value fails at the first use, maybe hours after the deploy.
- **Read at startup.** The program reads and parses all variables before it does any work. This is **fail fast**: a wrong setting stops the process immediately, with a clear message. A container that does not start after a deploy is easy to see and easy to roll back.

A good startup check also reports all errors together. With five wrong variables, five restarts that show one error each waste time.

Sources:

- [The Twelve-Factor App, "III. Config"](https://12factor.net/config): store config in the environment.
- [The Twelve-Factor App, "V. Build, release, run"](https://12factor.net/build-release-run): one build, and a release is the build plus the config.
- [Jim Shore, "Fail Fast"](https://martinfowler.com/ieeeSoftware/failFast.pdf), IEEE Software: why an immediate, visible failure is cheaper.
- [caarlos0/env](https://github.com/caarlos0/env): the library that parses environment variables into a Go struct.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| A config file for each environment in the image | The build includes `production.yaml` | One image for each environment. Secrets can get into git |
| Read each variable where the code needs it | A call to `os.Getenv`, like `process.env.X` in the middle of the code | A wrong value fails at first use. Nobody has the full list of settings |
| Small helper functions | `config.String(key, fallback)` and `config.Required(key)`. Our first version | Only strings. Each caller parses numbers. One error for each start |
| A typed struct, loaded one time at startup | Each field has a tag with the variable name and the default. One call parses all fields | A library, and reflection |

## Edge cases

| What if | What happens |
| --- | --- |
| A required variable is missing | The process stops at startup. The message names the variable and points to `.env.example` |
| Three variables are wrong | One message names all three, so the operator fixes them in one try |
| A value has the wrong type, for example `PORT=abc` | The library reports the Go field name. Our code maps it back to the variable name, because the operator sets the variable |
| An optional variable is empty, for example `SENTRY_DSN` | The feature is off. Error reports and the trace export turn on only when their variable is set |
| The settings are correct but the database is down | The process starts, because the database pool connects on first use. The readiness check answers 503 until the database answers |
| Somebody adds a new setting | It goes into the struct with a tag, and into `.env.example` with a comment. Compose too, if a container needs it |

## Decision

Settings come only from the environment. Each process (`api`, `worker`, `migrate`) describes its settings as a struct that embeds a shared `Base`. It calls `config.Load` one time at startup. A wrong environment stops the process with one message that names each wrong variable.

We started with the two helper functions. We replaced them when the settings got types (a port, a log level) and when one message for all errors became important.

This would be wrong for a setting that must change while the process runs, for example a feature flag. The process reads the environment one time, so such a setting needs a different source. There is no ADR for the rule. The library is in [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `.env.example`: each variable with a comment. The local setup copies it to `.env`.
2. `core/internal/platform/config/config.go`, `Base` and `Load`: the shared settings and the one message for all errors.
3. `core/internal/platform/config/config_test.go`, `TestLoadReportsAllErrors`: three wrong variables, one error.
4. `core/cmd/migrate/main.go`, `migrateConfig`: a small real example of a struct that embeds `Base`.

## Check yourself

1. Why is "the container does not start after the deploy" better than "the first request fails"?
2. Why does the error message name `DATABASE_URL` and not the Go field `DatabaseURL`?
3. The settings are valid, but the database is down at startup. Why does the process start, and is that a break of the fail-fast rule?

<details><summary>Answers</summary>

1. The failed start happens during the deploy, when a person or a pipeline watches and the old version can stay. The failed request happens later, for a real user, maybe at night. The cause is then far from the effect.
2. The person who fixes the problem sets environment variables and does not read Go code. The message must use the name that this person can change.
3. A wrong setting is a permanent error: a restart does not fix it, so the process must stop. A database that is down is a temporary state. The process starts and reports "not ready", and the proxy sends no traffic until the database answers. Fail fast is for errors that waiting cannot fix.

</details>
