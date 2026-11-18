---
name: find-solution
description: Find the best solution to a non-trivial problem with a dialectical loop of two agents, blind-solver (thesis) and solution-discriminator (antithesis), applying the first law of dialectics (quantity into quality) and Weinberg's Rule of Three. Use when the user runs /find-solution, or when choosing an approach/architecture, diagnosing a bug with several plausible causes, or weighing trade-offs.
---

# Find Solution: Dialectics + Rule of Three

## Principles

- **Rule of Three (Weinberg, *The Secrets of Consulting*):** fewer than 3 distinct solutions means you haven't thought enough. If you can't name 3 ways a plan fails, you don't understand it.
- **Quantity → quality (1st law of dialectics):** keep accumulating distinct candidates and constraints, round by round, until a qualitative leap: a candidate that survives the attack.
- **Unity and struggle of opposites:** the solver (thesis) and discriminator (antithesis) are kept apart. Their conflict produces the synthesis.

## Procedure

1. **Frame the problem.** Write a neutral problem statement: goal, context (relevant files, systems), and hard constraints. **Do not include your own preferred solution.** It would un-blind the solver.

2. **Round N** (start N=1, keep a running `constraints` list and a `candidates` ledger):
   a. Spawn `blind-solver` (Agent tool, `subagent_type: "blind-solver"`) with the problem statement + accumulated `constraints` only. Never pass it earlier solutions or critiques.
   b. Spawn `solution-discriminator` with the problem statement, constraints, this round's solutions, and the surviving (`viable`/`strong`) candidates from earlier rounds.
   c. Add its new constraints to `constraints`. Record every candidate and verdict in the ledger. If it proposed a synthesis, include it in the next discriminator call as a candidate.

3. **Stop when any of these holds:**
   - a `strong` candidate exists (qualitative leap), or
   - the round brought nothing qualitatively new, or
   - 3 rounds are done.
   Otherwise go to the next round.

4. **Report to the user:**
   - A table of all candidates: name, essence, verdict, key failure mode.
   - **Winner** and why it beats the runner-up.
   - At least 3 risks of the winner, each with a mitigation.
   - What was rejected and why (one line each).
   - Which stop criterion fired and how many rounds ran.

Don't implement anything inside this skill. It ends with a recommendation. Implementation follows the normal workflow once the user agrees.
