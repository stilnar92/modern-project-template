# 0016 — Alive or ready: when to restart a process and when to stop its traffic

From PR #105 · 12.11.2026 · Topics: health checks, operations

## Problem

Something outside the API process must know its state: Docker, a proxy, later an orchestrator. It asks the process over HTTP. The API needs the database for almost each request, and the database can stop or become slow. We must decide what the process answers then.

## Predict first

The process has one check, and the check also tests the database. The database stops for one minute. The platform restarts each process whose check fails. What happens to the API, and does it help?

## Theory

A platform asks a process two different questions:

- **Liveness:** "Is the process broken in a way that only a restart fixes?" Examples: a deadlock, a loop that never ends. A failed answer means "restart me".
- **Readiness:** "Can the process serve requests now?" Examples: it is still starting, or a dependency does not answer. A failed answer means "do not send me traffic, and do not restart me".

The two questions have different actions, so they need different checks. A restart does not repair a database. If the liveness check tests the database, a database failure makes the platform restart each healthy copy of the service, again and again. When the database comes back, the copies are in the middle of a restart.

So the rule is: the liveness check tests only the process. The readiness check tests the dependencies without which the process cannot do its work. Each check must be fast and must have a timeout, or a stopped dependency also stops the check.

Kubernetes names the two checks liveness probe and readiness probe. Docker has one `HEALTHCHECK`, which marks a container as healthy or unhealthy.

Sources:

- [Liveness, Readiness, and Startup Probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/), the Kubernetes documentation: the definitions.
- [Configure Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/), the Kubernetes documentation: examples and the caution about wrong liveness probes.
- [Dockerfile reference: HEALTHCHECK](https://docs.docker.com/reference/dockerfile/#healthcheck): the one check of Docker.
- Nygard, "Release It!", second edition, chapter 5 "Stability Patterns": timeouts, and how a failure moves from one system to the next.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| One check, only the process | The answer is always 200 while the process runs | Traffic goes to a process that cannot reach the database |
| One check that also tests the database | A database failure fails the check | A healthy process looks broken. A restart does not help |
| Two checks | Liveness tests the process. Readiness tests the database | Two endpoints, and each caller must use the correct one |

## Edge cases

| What if | What happens |
| --- | --- |
| Postgres stops | `/healthz` stays 200. `/readyz` answers 503 with a Problem body. We checked this in the real stack |
| Postgres starts again | `/readyz` answers 200 again. Nobody restarted the API |
| The database accepts the connection but does not answer | The ping has a timeout of 2 seconds, so `/readyz` answers 503 quickly |
| The database is down when the API starts | The API still starts. The pool connects on first use, and `/readyz` reports the problem |
| A client reads the 503 body | It gets only "The service is not ready." The reason goes to the log |
| The checks arrive each few seconds | Their access log records go to the Debug level |
| The container image has no shell | Docker cannot run `curl`. The API binary has a `-healthcheck` flag that calls its own `/healthz` |

## Decision

We use two checks. `/healthz` is liveness: it does not touch the database. `/readyz` is readiness: it pings the database with a timeout. In the local stack, Docker calls only `/healthz`. `/readyz` is for the proxy and an orchestrator. The contract in `api/openapi.yaml` describes both.

This would be wrong for a process with no dependency: one check is then sufficient. It would also be wrong to add each dependency to `/readyz`. If an optional dependency fails, for example analytics, the API can still serve most requests and must stay ready.

Record: [backend standard 06](../../dev/backend-standard/06-observability.md). The libraries are in [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `core/internal/platform/health/health.go`: `GetHealth` and `GetReady`, with the ping timeout.
2. `core/internal/platform/health/health_test.go`, `TestGetReady`: the test closes a database pool and expects 503.
3. `core/cmd/api/main.go`: `runHealthcheck`, and the comment about the pool in `run`.
4. `deploy/compose.yaml`: the `healthcheck` of the `api` service.

## Check yourself

1. Postgres stops. What does each check answer, and what does Docker do with the API container?
2. Why does the ping in `/readyz` have a timeout?
3. A colleague wants to add the email server to `/readyz`. Which question do you ask before you agree?

<details><summary>Answers</summary>

1. `/healthz` answers 200, so Docker keeps the container as healthy. `/readyz` answers 503. A caller that reads `/readyz` stops the traffic until the database answers again.
2. A database that hangs gives no error. Without a timeout the check also hangs, and the caller cannot tell "not ready" from "slow". With the timeout the answer is a 503 in 2 seconds.
3. "Can the API do its main work without the email server?" If it can, the email server must not be in the readiness check. One optional dependency would then take all traffic away from a process that works.

</details>
