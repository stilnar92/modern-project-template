# 22 — Accessibility (a11y)

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### A11y is UI architecture

Source: [A11Y is UI Architecture](https://dev.to/artalar/a11y-is-ui-architecture-2ibn) (Artyom / @artalar_dev).

- **A11y is architecture, not compliance** — it addresses limitations everyone has sometimes (one hand busy, glare, broken mouse, eye strain); building to it produces more robust UI for all users.
- **WAI-ARIA roles as the component vocabulary** — name components by *behavior*, not appearance; drop informal names (modal, toast, snackbar) in favor of the role taxonomy:
  - `dialog` — implies backdrop, focus lock, descriptive attributes;
  - `alert` — autonomous notification demanding attention;
  - `alertdialog` — notification requiring a response;
  - `status` / `log` / `marquee` — non-urgent updates.
  Feeds the glossary (section 03, not written yet): the role taxonomy is ubiquitous language for design–dev–agent communication.
- **ARIA Authoring Practices Guide (APG) is required reading** — proven interaction patterns (keyboard, focus, roles) for nearly every UI scenario; components follow APG patterns instead of inventing behavior.
- **Focus management and keyboard support are part of the component's definition**, not a polish pass — the role dictates them.
- **ARIA-first taxonomy reduces design decisions** — the component set and its behaviors are largely pre-decided by the spec; teams argue functional intent, not implementation.
