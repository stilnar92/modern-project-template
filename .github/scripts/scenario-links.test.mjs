// Tests for scenario-links.mjs. Run: mise run test:scripts
import assert from "node:assert/strict";
import { test } from "node:test";
import { missingScenarios, scenarioNames, tableRows } from "./scenario-links.mjs";

const spec = `## ADDED Requirements

### Requirement: Public page

#### Scenario: Visitor opens the page

- GIVEN a published page

#### Scenario:  Unknown slug
`;

const body = (rows) => `## Summary

Text.

## Scenarios → checks

| Scenario | Test or check |
| --- | --- |
${rows}

## Evidence

| Scenario | Test or check |
| --- | --- |
| Unknown slug | not in the right section |
`;

test("scenario names come from the spec delta", () => {
  assert.deepEqual(scenarioNames(spec), ["Visitor opens the page", "Unknown slug"]);
});

test("only the rows of the Scenarios → checks table count", () => {
  assert.deepEqual(tableRows(body("| A | b |")), [["A", "b"]]);
});

test("a scenario without a filled row is missing", () => {
  const names = scenarioNames(spec);
  assert.deepEqual(missingScenarios(names, body("| | |")), names, "the empty template row");
  assert.deepEqual(missingScenarios(names, body("| Visitor opens the page | `page_test.go` |\n| Unknown slug | — |")), ["Unknown slug"]);
  assert.deepEqual(
    missingScenarios(names, body("| **Visitor opens the page** | `TestPublicPage` |\n| unknown slug | e2e `site.spec.ts` |")),
    [],
    "markup and case do not matter",
  );
  assert.deepEqual(missingScenarios(names, ""), names, "no table");
});

test("one row covers only the scenario with the same name", () => {
  assert.deepEqual(missingScenarios(["Valid token", "Invalid token"], body("| Invalid token | `TestInvalid` |")), ["Valid token"]);
});
