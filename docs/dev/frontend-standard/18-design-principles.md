# 18 — Design Principles

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D8** — responsive by default: every component and screen is designed and reviewed across the full width range from its first version. Desktop-only is never an acceptable initial state.
- **D14** — fluid-first responsiveness (source: [The Too Early Breakpoint](https://ishadeed.com/article/too-early-breakpoint/), Ahmad Shadeed):
  - layouts must be correct at **every** viewport width, not only at named breakpoints;
  - in-between zones (split-screen, tablets, resized windows, link previews) are first-class usage, not edge cases;
  - breakpoints are driven by content, not device classes;
  - never switch to the mobile layout earlier than the content forces it;
  - prefer intrinsic layout (Grid/Flexbox, container queries, content-driven sizing) over viewport breakpoints.
