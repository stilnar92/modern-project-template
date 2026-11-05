# AI-First Frontend Standard

This is a reusable, product-agnostic documentation standard for building frontend projects with AI-first development: specs and documentation are written **before** the code, and both humans and AI agents treat them as the source of truth for every decision — architecture, dependencies, conventions, and process. Development itself is spec-driven and runs through a configured AI harness with reusable skills and standard agent workflows.

**Stack baseline:** React + TypeScript (mandatory). All other choices are defined and justified within these documents.

## How it's organized

Each area of the standard lives in its own numbered file. The master list of all areas — with scope descriptions and completion status — is in [00-overview.md](00-overview.md). Start there.

## How to use it

1. Read `00-overview.md` to see the full map.
2. Areas are written one at a time, top-down; a file exists only once its area status moves past `TODO`.
3. When starting a new project, this standard is applied as-is; project-specific deviations must be recorded as decisions, never made silently.
