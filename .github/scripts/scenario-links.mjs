#!/usr/bin/env node
// Each spec scenario of a feat/<change-id> branch needs a test or check in
// the "Scenarios → checks" table of the PR description (AI standard 03,
// step 2). Run: mise run check:scenarios. CI runs it on each PR change.
//
// The branch comes from GITHUB_HEAD_REF in CI, else from git. The PR body
// comes from the GitHub event in CI, else from "gh pr view".
import { execFileSync } from "node:child_process";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

// The scenario names in a spec delta: "#### Scenario: <name>".
export function scenarioNames(spec) {
  return [...spec.matchAll(/^####\s+Scenario:\s*(.+?)\s*$/gm)].map((m) => m[1]);
}

const norm = (s) => s.replace(/[`*_]/g, "").replace(/\s+/g, " ").trim().toLowerCase();

// The rows of the "Scenarios → checks" table: [scenario cell, check cell].
export function tableRows(body) {
  const section = body.split(/^##\s+/m).find((s) => /^Scenarios\s*(→|->)\s*checks/i.test(s)) ?? "";
  return section
    .split("\n")
    .filter((line) => /^\s*\|/.test(line) && !/^\s*\|[\s|:-]+\|\s*$/.test(line))
    .map((line) => line.trim().replace(/^\||\|$/g, "").split("|").map((cell) => cell.trim()))
    .filter((cells) => cells.length >= 2 && norm(cells[0]) !== "scenario");
}

const EMPTY = /^(|-|—|todo|tbd|n\/a)$/i;

// The scenarios that have no row with a test or check. The row must name the
// whole scenario: "Invalid token" does not cover "Valid token".
export function missingScenarios(scenarios, body) {
  const rows = tableRows(body).map(([scenario, check]) => [norm(scenario), check]);
  return scenarios.filter((name) => !rows.some(([cell, check]) => cell === norm(name) && !EMPTY.test(check)));
}

function specFiles(dir) {
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) return specFiles(p);
    return e.name === "spec.md" ? [p] : [];
  });
}

function sh(cmd, args) {
  return execFileSync(cmd, args, { encoding: "utf8", stdio: ["ignore", "pipe", "inherit"] }).trim();
}

function prBody() {
  const eventPath = process.env.GITHUB_EVENT_PATH;
  if (eventPath && existsSync(eventPath)) {
    return JSON.parse(readFileSync(eventPath, "utf8")).pull_request?.body ?? "";
  }
  return sh("gh", ["pr", "view", "--json", "body", "--jq", ".body"]);
}

function main() {
  const branch = process.env.GITHUB_HEAD_REF || sh("git", ["branch", "--show-current"]);
  if (!branch.startsWith("feat/")) {
    console.log(`${branch} is not a feat/ branch: no spec scenarios to check.`);
    return 0;
  }
  const id = branch.slice("feat/".length);
  const files = specFiles(path.join("openspec", "changes", id, "specs"));
  if (files.length === 0) {
    console.error(`No spec deltas in openspec/changes/${id}/specs/. A feat/ branch needs its OpenSpec change (AI standard 01).`);
    return 1;
  }
  const scenarios = files.flatMap((f) => scenarioNames(readFileSync(f, "utf8")));
  const missing = missingScenarios(scenarios, prBody());
  if (missing.length > 0) {
    console.error(
      'These scenarios have no test or check in the "Scenarios → checks" table of the PR description:\n' +
        missing.map((m) => `- ${m}`).join("\n"),
    );
    return 1;
  }
  console.log(`All ${scenarios.length} scenarios have a test or check.`);
  return 0;
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) process.exit(main());
