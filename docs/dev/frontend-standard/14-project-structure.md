# 14 — Project Structure

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

- **The UI kit and business components live in separate trees** (user decision, [12](12-architecture.md)): the kit under a dedicated root (e.g. `src/ui/` with `atoms/`, `molecules/`, `organisms/` per the atomic taxonomy, [20](20-component-system.md)); business components inside their feature (e.g. `src/features/<feature>/components/`). Templates/pages materialize in routing/feature layers, not as a global folder.
