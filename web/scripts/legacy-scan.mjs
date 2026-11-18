// Fails when modern-web-guidance legacy-scan finds a legacy web pattern in
// the app and package sources (frontend standard 10). The tool itself exits
// with 0 even when it finds something, so this script reads its JSON output.
//
// An exception covers one file and one rule and has a reason. Add it to
// legacy-scan.allow.json, for example a fallback that a guide requires:
//   [{ "file": "apps/dashboard/src/x.css", "ruleId": "<rule>", "reason": "<why>" }]
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";

const roots = ["apps", "packages"];
const allow = JSON.parse(
  readFileSync(new URL("../legacy-scan.allow.json", import.meta.url), "utf8"),
);

const isAllowed = (root, file, ruleId) =>
  allow.some((a) => a.ruleId === ruleId && a.file === `${root}/${file}`);

let failed = 0;
for (const root of roots) {
  const out = execFileSync("modern-web-guidance", ["legacy-scan", root, "--json"], {
    encoding: "utf8",
    // The tool sends usage data to Google by default.
    env: { ...process.env, DISABLE_TELEMETRY: "1" },
  });
  const report = JSON.parse(out);
  for (const guide of report.guides) {
    for (const rule of guide.rules) {
      for (const f of rule.findings) {
        if (isAllowed(root, f.file, rule.ruleId)) continue;
        failed++;
        console.error(`${root}/${f.file}:${f.line} [${rule.ruleId}] ${rule.summary}`);
        console.error(`  Guide: pnpm exec modern-web-guidance retrieve ${guide.guideId}`);
      }
    }
  }
}

if (failed > 0) {
  console.error(`\n${failed} legacy pattern(s). Use the modern pattern from the guide.`);
  console.error("If the code is a required fallback, add an exception to legacy-scan.allow.json.");
  process.exit(1);
}
console.log("legacy-scan: no legacy patterns.");
