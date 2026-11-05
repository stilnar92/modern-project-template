# 0003 — Analytics: events in the core, a self-hosted tool as a sink

Date: 2026-10-28
Status: accepted

## Context

We need a self-hosted analytics system. The project has two different kinds of analytics:

- **for customers:** reports inside the product about their own data. This is part of the product that customers pay for;
- **for us:** activation, feature usage, retention, the success metrics of the product ([brief](../brief.md)). D13 of the [frontend standard](../frontend-standard/00-overview.md) also requires it: product analytics from day one.

Customer analytics links steps of the domain model into chains. It is isolated per account and can contain the personal data of the customer's own end users. For that data the customer is the controller and we are the processor. An external tool cannot do this: it does not have our model or account isolation, and we must not use the customer's end-user data for our own purposes.

## Decision

- **The event log is in the core.** The core writes an event to the Postgres `events` table in the same transaction as the action (outbox). This is the source of truth for both kinds of analytics.
- **The core builds the customer reports** from its own tables and shows them in the dashboard.
- **The self-hosted tool is a sink for "our" analytics.** A River job sends events to it through a sink adapter. If the tool is down, the events wait in the queue. To change the tool, we replace one adapter.
- **Only data about accounts and their users goes to the tool,** with the account ID as a hash. The personal data of the customers' end users never goes there.
- **The tool's JS snippet is only in the dashboard.** Public pages do not have it: they send events only to the core.
- **The tool runs on our own server,** next to the rest of the platform.
- **We select the tool during the foundation work.** Candidates: PostHog (funnels, retention, feature flags; self-hosting is heavy, check the terms), Umami or Plausible (light, but mostly web analytics), Matomo.

## Consequences

- Customer analytics does not depend on an external tool or its availability.
- We get product analytics ready-made: we do not write funnels and retention ourselves.
- Each event passes the filter "what can go outside" in the sink adapter. A new event field needs a decision: does it go to the tool or not.
- One more service to operate: backups, updates, monitoring.
