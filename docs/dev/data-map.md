# Personal data map

This document lists the personal data that the system holds, where, and for how long. It is an engineering map. Legal documents give the roles, the consent text and the regulatory obligations. This map gives the table or field that holds each kind of data and the time when the system deletes it. If the two differ, the legal document is correct, and you correct this map.

Update this map in the same PR as a migration that adds personal data or a new storage place ([development process](dev-process.md#on-events)).

## Roles

<!-- One row for each kind of person whose data the system holds. -->

| Whose data | Who is the controller | Our role | Basis |
| --- | --- | --- | --- |
| <!-- e.g. Users of the product --> | | | |

## Data by entity

<!-- One row for each entity or field. Name the domain entity from CONTEXT.md, not a physical table. -->

| Entity / field | Personal data? | Where stored | Retention | Deletion trigger |
| --- | --- | --- | --- | --- |
| <!-- e.g. User: name, email --> | | <!-- Postgres, S3 --> | | |

**Not collected:** <!-- the data that the system must never store -->

## Deletion

- Delete personal data physically when its retention period ends. Do not soft-delete it ([backend standard 03](backend-standard/03-data.md#schema)).
- A scheduled job deletes the data. It can use the system database role (`BYPASSRLS`), which is one of the limited, audited uses of that role ([ADR 0009](adr/0009-account-isolation-rls.md#decision)).
- Keep a deletion log: what was deleted, when, and why. Then you can answer a data-subject request or an audit.

## Data residency

<!-- Which countries or regions can hold the data: database, object storage, external providers. -->

## Consent

<!-- What consent the system collects, per what scope, and where it keeps the consent log. -->

## Open points

<!-- Questions that a person must answer before the system collects real data. -->

- [ ]
