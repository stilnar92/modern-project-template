# 12 — Architecture

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D3** — strict separation of UI and business logic is mandatory: components render; business logic lives in hooks/services and is testable without the UI. Crossing this boundary is never justified by speed.

### UI kit vs business components (user decision)

Two distinct component layers, never mixed:

- **UI kit** — product-agnostic, presentational: knows tokens ([19](19-design-tokens.md)), variants (D6), headless behavior ([20](20-component-system.md)); imports **nothing** from the domain (no API types, no stores, no feature code). Organized by the atomic taxonomy.
- **Business components** — feature-owned: compose kit components and carry domain meaning; their logic lives in hooks/services per D3. Not forced into atomic naming.
- **Dependency direction:** features → kit, never the reverse. A kit component importing domain code is a boundary violation — mechanically enforced ([13](13-architecture-enforcement.md)).
