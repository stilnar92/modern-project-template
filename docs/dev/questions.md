# Open questions

One entry point to the open questions of the project. Each question stays in the document that owns it, so this page links to those sections and does not copy them. Only the infrastructure choices and the deferred setup below have no other home. An answer goes into its source document, an ADR, or an OpenSpec change. Then remove it here.

## Where the questions are

| Area | Section | Kind |
| --- | --- | --- |
| Domain model | [CONTEXT.md](../../CONTEXT.md) | Slice specs settle them |
| Non-functional targets | [nfr.md, What still needs a real decision](nfr.md#what-still-needs-a-real-decision) | Measurements |
| Personal data | [data-map.md, Open points](data-map.md#open-points) | Legal and schema |
| Tech stack | [Infrastructure and operations](tech-stack.md#infrastructure-and-operations) | Infrastructure |

## Infrastructure choices

These choices are expensive to change. The owner makes each one with a decision card ([AI standard 05](ai-standard/05-human-in-the-loop.md#decision-card)), and an ADR records it.

| Choice | Options | Source |
| --- | --- | --- |
| Product analytics tool (sink) | PostHog, Umami, Plausible, Matomo | [ADR 0003](adr/0003-analytics-outbox-and-sinks.md) |
| Error tracker | Sentry (own server) or GlitchTip | [Backend standard 06](backend-standard/06-observability.md) |
| Backend for traces, metrics and logs | For example Prometheus, Grafana and Loki, or one Grafana LGTM container. No trace backend is named yet | [Tech stack](tech-stack.md#infrastructure-and-operations) |
| Hosting of the core and the dashboard | Server, region, and how the dashboard gets the session cookie: same origin or CORS | [ADR 0005](adr/0005-api-surfaces-and-ownership.md), [deploy/README.md](../../deploy/README.md) |
| Platform secrets | The CI secret store or SOPS | [Tech stack](tech-stack.md#infrastructure-and-operations) |
| Rules for Astro pages | Which frontend standard rules apply to Astro, React islands or plain Astro, server-only error tracking for the public pages | [Tech stack](tech-stack.md#frontend), [site README](../../web/apps/site/README.md) |
| Design tokens and Tailwind | Token names and how the dashboard theme and the public pages share them | [Frontend standard 19](frontend-standard/19-design-tokens.md) |

## Deferred setup

Setup work that is not done yet. None of it needs a domain decision. A library choice goes to the owner first, with the researched options ([ADR 0012](adr/0012-infrastructure-libraries.md)).

- **Waits for the login flow** (a product decision): the real login method, a session endpoint in the contract, and the route guard of the dashboard ([frontend standard 25](frontend-standard/25-routing.md#guard)).
- **Waits for the first product slice:** agent evals, and a sink for the analytics tool after its choice ([ADR 0003](adr/0003-analytics-outbox-and-sinks.md)).
- **Waits for the design:** the brand colors of `web/packages/tokens` (placeholders now).
- **Waits for a Codex session:** a check of the Codex hooks and `.codex/config.toml`.
- **Waits for hosting:** the trace backend for OTLP, the alert rules on the queue metrics, the error tracker server.
