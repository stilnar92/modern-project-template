# 37 — Task Playbook

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.
**Group:** B. AI Development System (numbered 37 to avoid renumbering existing files).

---

## Collected points

### When a task arrives (from the GitHub board)

1. **Understand before touching code:**
   - what is actually being asked, and *why* — the problem behind the ticket, not just its literal text;
   - acceptance criteria — if they're missing or vague, that's resolved first (ties into spec-driven development, [05](05-spec-driven-development.md));
   - unclear points become questions to the ticket author, not assumptions.
2. **Understand the modules involved:**
   - map which modules/features the task touches and their boundaries ([12](12-architecture.md));
   - trace the affected code paths before changing them;
   - identify blast radius: what else depends on the code being changed.
3. **Verify assumptions:**
   - for bugs — reproduce first; no fix without a reproduction;
   - check the current actual behavior against what the ticket claims;
   - check for existing utilities/patterns that already solve part of the task — don't build what exists.
4. **Consider an MVP slice:**
   - if the task is big, cut the smallest version that delivers value / validates the approach, ship or demo it, then iterate;
   - the slice boundaries are agreed (in the ticket/spec), not silently chosen.
5. **Do → check → test:**
   - implement following the agent workflow ([08](08-agent-workflows.md));
   - test what the acceptance criteria state, plus the affected-module boundaries from step 2;
   - verify end-to-end in the running app, not only via unit tests.
6. **Done means verified:**
   - all quality gates pass ([17](17-quality-gates.md));
   - acceptance criteria demonstrably met;
   - ticket updated with what was done and any deviations from the plan.
