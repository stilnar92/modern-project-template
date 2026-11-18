#!/usr/bin/env node
// One entry for all Claude Code hooks (AI standard 02, "Hooks").
// It reads the hook event from stdin and decides.
// Exit 0: allow. Exit 2: block (PreToolUse) or report to the agent
// (PostToolUse). The message goes to stderr.
import { execFileSync } from "node:child_process";
import { existsSync, readdirSync, readFileSync, realpathSync } from "node:fs";
import path from "node:path";

const event = JSON.parse(readFileSync(0, "utf8") || "{}");
const root = realpathSync(process.env.CLAUDE_PROJECT_DIR || event.cwd || process.cwd());

function block(message) {
  process.stderr.write(message + "\n");
  process.exit(2);
}

function git(args, cwd = event.cwd || root) {
  try {
    return execFileSync("git", args, { cwd, encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim();
  } catch {
    return "";
  }
}

// The real path: on macOS /var is a link to /private/var, and git prints
// the real form. The file itself may not exist yet, so resolve its folder.
function real(p) {
  let dir = path.dirname(p);
  while (!existsSync(dir) && dir !== path.dirname(dir)) dir = path.dirname(dir);
  return path.join(realpathSync(dir), path.relative(dir, p));
}

// The worktree of a file in a tool call and the path inside it. An agent
// can edit a file in another worktree, so the rules use the file's worktree.
function locateFile(file) {
  if (!file) return { top: root, rel: "" };
  const abs = real(path.resolve(event.cwd || root, file));
  let dir = path.dirname(abs);
  while (!existsSync(dir)) dir = path.dirname(dir);
  const top = git(["rev-parse", "--show-toplevel"], dir) || root;
  return { top, rel: path.relative(top, abs).split(path.sep).join("/") };
}

const tool = event.tool_name;
const input = event.tool_input || {};
const isEdit = ["Edit", "Write", "MultiEdit", "NotebookEdit"].includes(tool);
const { top: fileTop, rel: file } = locateFile(input.file_path || input.notebook_path);

// --- Rule 1: shared files for the orchestrator only --------------------
// A subagent call has agent_id. Shared files define the work of all agents,
// so only the main agent changes them.
const SHARED = [
  /^api\/openapi\.yaml$/,
  /^core\/migrations\//,
  /(^|\/)events\.md$/,
  /(^|\/)CONTEXT\.md$/,
  /(^|\/)THESAURUS\.md$/,
  /^openspec\/specs\//,
];

function checkShared() {
  if (event.hook_event_name !== "PreToolUse" || !isEdit || !event.agent_id) return;
  if (SHARED.some((re) => re.test(file))) {
    block(
      `Blocked: ${file} is a shared file. Only the orchestrator changes it ` +
        `(AI standard 02). Report the needed change to the orchestrator.`,
    );
  }
}

// --- Rule 2: no code without an approved plan ---------------------------
// Code is a source file in core/ or web/. Allowed on a chore/, fix/ or
// docs/ branch (tooling and bug fixes need no proposal), or on feat/<id>
// when openspec/changes/<id>/proposal.md has the line "Status: approved".
const CODE = /^(core|web)\/.+\.(go|sql|ts|tsx|astro|css)$/;

// One test for "approved", used by the edit check and the session start, so
// they always agree. Plain "Status: approved" and bold Markdown
// ("**Status:** approved", "**Status: approved**") pass. "Status: draft" fails.
const APPROVED = /^(\*\*)?Status:(\*\*)?\s*approved(\*\*)?\s*$/im;
const isApproved = (proposal) => APPROVED.test(readFileSync(proposal, "utf8"));
const NO_PLAN_NEEDED = /^(chore|fix|docs)\//;

function checkPlan() {
  if (event.hook_event_name !== "PreToolUse" || !isEdit || !CODE.test(file)) return;
  if (/(^|\/)(node_modules|dist)\//.test(file)) return;
  const branch = git(["branch", "--show-current"], fileTop);
  if (branch === "main" || branch === "") {
    block(`Blocked: do not change code on "${branch || "detached HEAD"}". Create a branch first (feat/<change-id>, fix/..., chore/...).`);
  }
  if (NO_PLAN_NEEDED.test(branch)) return;
  const id = branch.replace(/^feat\//, "");
  const proposal = path.join(fileTop, "openspec", "changes", id, "proposal.md");
  if (!branch.startsWith("feat/") || !existsSync(proposal)) {
    block(
      `Blocked: no approved plan for branch "${branch}". Use feat/<change-id> with ` +
        `openspec/changes/<change-id>/proposal.md, or a chore/, fix/ or docs/ branch (AI standard 01).`,
    );
  }
  if (!isApproved(proposal)) {
    block(`Blocked: ${path.relative(fileTop, proposal)} has no "Status: approved" line. The owner approves the plan first.`);
  }
}

// --- Rule 3: no push to main --------------------------------------------
// An allow list, not a block list: a push passes only when it names the
// remote and each branch, and no branch is main. A push without an
// explicit branch is blocked, because the hook cannot know what it pushes
// (for example after "git checkout main" or with "git -C <other worktree>").

// Splits a shell command into simple commands and their words. It knows
// quotes and backslashes, and splits on ; & | and newlines outside quotes.
function commands(text) {
  const out = [[]];
  let word = null;
  let quote = "";
  let dropNext = false; // the next word is the target of a redirection
  const end = () => {
    if (word !== null && !dropNext) out[out.length - 1].push(word);
    if (word !== null) dropNext = false;
    word = null;
  };
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (quote) {
      if (c === quote) quote = "";
      else if (c === "\\" && quote === '"' && i + 1 < text.length) word += text[++i];
      else word += c;
    } else if (c === "'" || c === '"') {
      quote = c;
      word ??= "";
    } else if (c === "\\" && i + 1 < text.length) {
      word = (word ?? "") + text[++i];
    } else if (c === ">" || c === "<") {
      // A redirection (>, >>, 2>, <, >&2): the shell removes it and its
      // target before the command runs, so they are not arguments.
      if (word !== null && /^\d+$/.test(word)) word = null;
      end();
      while (text[i + 1] === ">" || text[i + 1] === "<") i++;
      if (text[i + 1] === "&") {
        i++;
        while (/[\d-]/.test(text[i + 1] ?? "")) i++;
      } else {
        dropNext = true;
      }
    } else if (/\s/.test(c) && c !== "\n") {
      end();
    } else if (c === ";" || c === "&" || c === "|" || c === "\n" || c === "(" || c === ")") {
      end();
      out.push([]);
    } else {
      word = (word ?? "") + c;
    }
  }
  end();
  return out.filter((words) => words.length > 0);
}

// Programs that run the next words as a command.
const WRAPPERS = ["env", "command", "exec", "sudo", "nice", "time", "nohup"];
const WRAPPER_VALUE_OPTIONS = {
  env: ["-u", "--unset", "-C", "--chdir"],
  sudo: ["-u", "--user", "-g", "--group", "-h", "--host", "-p", "--prompt", "-C", "--close-from", "-D", "--chdir", "-R", "--chroot", "-T", "--command-timeout"],
  nice: ["-n", "--adjustment"],
  time: ["-o", "--output", "-f", "--format"],
  exec: ["-a"],
};

// The index of the command word: it comes after VAR=value assignments,
// wrappers such as env and sudo, and their options.
function commandStart(words) {
  let g = 0;
  while (g < words.length) {
    if (/^[A-Za-z_][A-Za-z0-9_]*=/.test(words[g])) {
      g++;
      continue;
    }
    const wrapper = path.basename(words[g]);
    if (!WRAPPERS.includes(wrapper)) break;
    g++;
    while (g < words.length && words[g].startsWith("-")) {
      const option = words[g++];
      if (option === "--") break;
      if (WRAPPER_VALUE_OPTIONS[wrapper]?.includes(option)) g++;
    }
  }
  return g;
}

// The git subcommand of a simple command: its name, its index in words, and
// the folders of -C options. The subcommand is the first word after "git"
// and its options, so "git commit -m 'push to main'" is a commit, not a push.
function gitSubcommand(words) {
  const g = commandStart(words);
  if (g >= words.length || path.basename(words[g]) !== "git") return { name: "", index: -1, dirs: [] };
  const dirs = [];
  let i = g + 1;
  while (i < words.length && words[i].startsWith("-")) {
    if (words[i] === "-C" && i + 1 < words.length) dirs.push(words[i + 1]);
    // These options take the next word as their value.
    if (["-C", "-c", "--git-dir", "--work-tree", "--namespace"].includes(words[i])) i++;
    i++;
  }
  return { name: words[i] ?? "", index: i, dirs };
}

// All simple commands of a Bash call, also the commands inside
// bash -c "...", sh -c "..." and zsh -c "...".
function bashCommands(text) {
  const queue = commands(text);
  const all = [];
  while (queue.length > 0) {
    const words = queue.shift();
    all.push(words);
    const shell = words.findIndex((w) => /^(ba|z)?sh$/.test(path.basename(w)));
    if (shell < 0) continue;
    // The first word after short options that include "c" (-c, -lc, -ec).
    let i = shell + 1;
    let hasC = false;
    while (i < words.length && /^-[a-z]+$/.test(words[i])) {
      hasC = words[i].includes("c") || hasC;
      i++;
    }
    if (hasC && words[i]) queue.push(...commands(words[i]));
  }
  return all;
}

// A destination branch that is safe to push: a plain name, not main.
function safeDestination(refspec) {
  const ref = refspec.replace(/^\+/, "");
  const dest = (ref.includes(":") ? ref.split(":").pop() : ref).replace(/^refs\/heads\//, "");
  // $VAR or `cmd` is unknown until the shell runs: block it.
  if (/[$`]/.test(refspec)) return false;
  return dest !== "" && !/^(main|master|HEAD)$/.test(dest) && !/[*?[]/.test(dest) && !dest.startsWith("refs/");
}

// Options of git push that take the next word as their value.
const PUSH_VALUE_OPTIONS = ["--repo", "--receive-pack", "--exec", "-o", "--push-option"];

function checkPush() {
  if (event.hook_event_name !== "PreToolUse" || tool !== "Bash") return;
  for (const words of bashCommands(String(input.command || ""))) {
    const sub = gitSubcommand(words);
    if (sub.name !== "push") continue;
    const at = sub.index;
    const positional = [];
    let repoOption = false;
    let bulk = false;
    for (let i = at + 1; i < words.length; i++) {
      const w = words[i];
      if (!w.startsWith("-")) positional.push(w);
      else if (/^--(all|mirror|branches|tags)$/.test(w)) bulk = true;
      else if (w.startsWith("--repo")) {
        repoOption = true;
        if (w === "--repo") i++;
      } else if (PUSH_VALUE_OPTIONS.includes(w)) i++;
    }
    const refspecs = repoOption ? positional : positional.slice(1);
    if (bulk || refspecs.length === 0 || !refspecs.every(safeDestination)) {
      block(
        "Blocked: agents do not push to main. Push with an explicit branch name: " +
          "git push -u origin <branch>. Then open a PR (AI standard 02).",
      );
    }
  }
}

// --- Rule 6: a subagent does not merge ------------------------------------
// The orchestrator merges after the review (AI standard 01). A subagent
// reports its result and stops: "gh pr merge" and merge calls of the GitHub
// API ("gh api .../pulls/N/merge", "gh api .../merges") are blocked for it.
function checkSubagentMerge() {
  if (event.hook_event_name !== "PreToolUse" || tool !== "Bash" || !event.agent_id) return;
  for (const words of bashCommands(String(input.command || ""))) {
    const g = commandStart(words);
    if (path.basename(words[g] ?? "") !== "gh") continue;
    const args = words.slice(g + 1).filter((w) => !w.startsWith("-"));
    const merge = (args[0] === "pr" && args[1] === "merge") || (args[0] === "api" && args.some((w) => /\/merges?$/.test(w)));
    if (merge) block("Blocked: a subagent does not merge. Report the result to the orchestrator (AI standard 01).");
  }
}

// --- Rule 4: fast checks after an edit ---------------------------------
// Only fast checks (frontend standard D11): gofmt and go vet for Go, oxlint
// for TypeScript. A problem goes back to the agent, the edit stays.

// The path of a tool from mise.toml. The shell may not activate mise, so
// the hook asks mise for the pinned binary. Without mise, use PATH.
function which(name, cwd) {
  try {
    return execFileSync("mise", ["which", "-C", cwd, name], { encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] }).trim() || name;
  } catch {
    return name;
  }
}

// Runs a check. It returns "" when the check passes or the tool is not
// installed, and the output when the check fails or does not finish in time.
function run(cmd, args, cwd, timeout = 60_000) {
  try {
    execFileSync(cmd, args, { cwd, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"], timeout: Math.max(timeout, 1) });
    return "";
  } catch (err) {
    if (err.code === "ENOENT") return ""; // the tool is not installed: skip
    if (err.code === "ETIMEDOUT") return `${path.basename(cmd)}: the check did not finish in time.\n`;
    return (`${err.stdout || ""}${err.stderr || ""}`.trim() || String(err.message)) + "\n";
  }
}

// A web tool from web/node_modules. A worktree may have no node_modules, so
// look in the worktree, then in the project, then next to this hook file.
function webTool(name, top) {
  const project = path.resolve(import.meta.dirname, "..", "..");
  return [top, root, project].map((d) => path.join(d, "web", "node_modules", ".bin", name)).find(existsSync);
}

const GO_FILE = /^core\/.+\.go$/;
const TS_FILE = /^web\/.+\.(ts|tsx)$/;

// The fast checks on files (paths from the worktree top). It returns the
// problems as text, or "" when there are none.
function fastChecks(top, files, deadline) {
  const left = () => deadline - Date.now();
  const goFiles = files.filter((f) => GO_FILE.test(f));
  const tsFiles = files.filter((f) => TS_FILE.test(f) && !f.endsWith(".d.ts") && !/(^|\/)(node_modules|dist)\//.test(f));
  let problems = "";
  if (goFiles.length > 0) {
    // gofmt -l exits 0 and prints the name of each unformatted file.
    let listed = "";
    try {
      listed = execFileSync(which("gofmt", top), ["-l", ...goFiles], {
        cwd: top,
        encoding: "utf8",
        stdio: ["ignore", "pipe", "pipe"],
        timeout: Math.max(left(), 1),
      }).trim();
    } catch (err) {
      if (err.code !== "ENOENT") listed = `${err.stderr || err.message}`.trim();
    }
    if (listed) problems += `gofmt: these files are not formatted (run gofmt -w):\n${listed}\n`;
    const packages = [...new Set(goFiles.map((f) => "./" + path.relative("core", path.dirname(f))))];
    problems += run(which("go", top), ["vet", ...packages], path.join(top, "core"), left());
  }
  if (tsFiles.length > 0) {
    const abs = tsFiles.map((f) => path.join(top, f));
    const web = path.join(top, "web");
    const oxlint = webTool("oxlint", top);
    if (oxlint) problems += run(oxlint, ["--deny-warnings", ...abs], web, left());
    const oxfmt = webTool("oxfmt", top);
    if (oxfmt) {
      const found = run(oxfmt, ["--check", ...abs], web, left());
      if (found) problems += `oxfmt: these files are not formatted (run pnpm format in web/):\n${found}`;
    }
  }
  return problems.trim();
}

function checkAfterEdit() {
  if (event.hook_event_name !== "PostToolUse" || !isEdit) return;
  const problems = fastChecks(fileTop, [file], Date.now() + 60_000);
  if (problems) {
    process.stderr.write(`Fast checks found problems in ${file}:\n${problems}\n`);
    process.exit(2);
  }
}

// --- Rule 5: no commit without checks ----------------------------------
// Before "git commit", run the fast checks on each file that the commit can
// contain, and look for secrets. A hook that passes its time limit does not
// block the tool call, so this rule has its own, shorter time budget.
const COMMIT_BUDGET_MS = 45_000;

// Files that a commit can contain: staged, changed (git commit -a) and new
// (git add -A && git commit). At PreToolUse time, "git add" has not run yet.
function commitFiles(top) {
  const lists = [
    ["diff", "--cached", "--name-only", "--diff-filter=ACMR"],
    ["diff", "--name-only", "--diff-filter=ACMR"],
    ["ls-files", "--others", "--exclude-standard"],
  ].map((args) => git(args, top).split("\n").filter(Boolean));
  const files = [...new Set(lists.flat())].filter((f) => existsSync(path.join(top, f)));
  return { files, untracked: lists[2] };
}

function secrets(top, untracked, deadline) {
  const gitleaks = which("gitleaks", top);
  const left = () => deadline - Date.now();
  const flags = ["--redact", "--no-banner", "--no-color", "--log-level", "warn", "--verbose"];
  let found = run(gitleaks, ["git", "--pre-commit", "--staged", ...flags, "."], top, left());
  found += run(gitleaks, ["git", "--pre-commit", ...flags, "."], top, left());
  for (const f of untracked.slice(0, 50)) found += run(gitleaks, ["dir", ...flags, f], top, left());
  return found.trim() ? `gitleaks found possible secrets:\n${found.trim()}` : "";
}

function checkCommit() {
  if (event.hook_event_name !== "PreToolUse" || tool !== "Bash") return;
  for (const words of bashCommands(String(input.command || ""))) {
    const sub = gitSubcommand(words);
    if (sub.name !== "commit") continue;
    const cwd = path.resolve(event.cwd || root, ...sub.dirs);
    const top = git(["rev-parse", "--show-toplevel"], cwd) || root;
    const deadline = Date.now() + COMMIT_BUDGET_MS;
    const { files, untracked } = commitFiles(top);
    const problems = [fastChecks(top, files, deadline), secrets(top, untracked, deadline)].filter(Boolean).join("\n");
    if (problems) block(`Blocked: the fast checks failed for this commit. Fix the problems, then commit again (AI standard 02).\n${problems}`);
    return;
  }
}

// --- Rule 8: no shell reads of local secret files ---------------------------
// Claude Code denies reading .env through permissions.deny, but the Codex CLI
// has no such list (AI standard 02, "Secrets"). This rule blocks the shell
// programs that print or load a .env file. "--env-file .env" of docker
// compose passes: it gives the file to the containers, not to the agent.
// .env.example stays readable. Like the other rules, it stops honest
// mistakes, not a script that opens the file itself.
const SECRET_FILE = /(^|\/)\.env(\.local|\.production|\.[^./]+\.local)?$/;
const READERS = new Set(["cat", "head", "tail", "less", "more", "grep", "egrep", "rg", "sed", "awk", "bat", "nl", "xxd", "od", "strings", "cut", "sort", "base64", "source", ".", "cp", "diff"]);

function checkSecretRead() {
  if (event.hook_event_name !== "PreToolUse" || tool !== "Bash") return;
  for (const words of bashCommands(String(input.command || ""))) {
    const at = words.findIndex((w) => !/^\w+=/.test(w));
    if (at < 0 || !READERS.has(path.basename(words[at]))) continue;
    if (words.slice(at + 1).some((w) => SECRET_FILE.test(w.replace(/^<+/, "")))) {
      block("Blocked: do not read .env files: they hold local secrets. Read .env.example for the names (AI standard 02, \"Secrets\").");
    }
  }
}

// --- Rule 7: task state after a context compaction ------------------------
// A summary keeps the work and loses the role and the reasons (AI standard
// 07). After a compaction or a resume, tell the orchestrator to read the
// progress file before it does anything. The text goes into the context.
function progressFiles() {
  const id = git(["branch", "--show-current"]).replace(/^feat\//, "");
  const own = path.join("openspec", "changes", id, "progress.md");
  if (id && existsSync(path.join(root, own))) return [own];
  const changes = path.join(root, "openspec", "changes");
  if (!existsSync(changes)) return [];
  // Archived changes are one level deeper (archive/<date-id>/), so they do not match.
  return readdirSync(changes)
    .map((d) => path.join("openspec", "changes", d, "progress.md"))
    .filter((f) => existsSync(path.join(root, f)));
}

// --- Rule 9: task state into the context at the start ---------------------
// A new session gets the phase of its branch (AI standard 01, 02): the agent
// knows what it may do before it reads the whole process.
function taskState() {
  const branch = git(["branch", "--show-current"]);
  if (!branch || /^(main|master)$/.test(branch)) {
    return `Branch: ${branch || "detached HEAD"}. Code edits are blocked here. Create a branch first: feat/<change-id> for a feature, chore/, fix/ or docs/ for other work (AI standard 01).`;
  }
  const feat = branch.match(/^feat\/(.+)$/);
  if (!feat) {
    return `Branch: ${branch}. Setup, fix or docs work: no proposal is necessary. Finish with "mise run check" and a PR (AI standard 01).`;
  }
  const dir = path.join("openspec", "changes", feat[1]);
  const proposal = path.join(root, dir, "proposal.md");
  if (!existsSync(proposal)) {
    return `Branch: ${branch}. Phase: plan. ${dir}/proposal.md does not exist: write it with /to-spec. Code edits stay blocked until the owner approves it.`;
  }
  if (!isApproved(proposal)) {
    return `Branch: ${branch}. Phase: plan. ${dir}/proposal.md waits for the owner's approval. Do not write code yet.`;
  }
  const tasks = path.join(root, dir, "tasks.md");
  const open = existsSync(tasks) ? (readFileSync(tasks, "utf8").match(/^\s*- \[ \]/gm) || []).length : 0;
  return `Branch: ${branch}. Phase: code. The proposal is approved. Open tasks in ${dir}/tasks.md: ${open}. Keep ${dir}/progress.md up to date for a long task.`;
}

function checkSessionStart() {
  if (event.hook_event_name !== "SessionStart") return;
  if (["startup", "clear"].includes(event.source)) {
    process.stdout.write(JSON.stringify({ hookSpecificOutput: { hookEventName: "SessionStart", additionalContext: taskState() } }));
    process.exit(0);
  }
  if (!["compact", "resume"].includes(event.source)) return;
  const files = progressFiles().map((f) => f.split(path.sep).join("/"));
  const text =
    files.length > 0
      ? `The context was ${event.source === "compact" ? "compacted" : "resumed"}. Before you do anything, read ${files.join(", ")}. ` +
        "Then check each subagent that it lists: a subagent can still run or be already finished (AI standard 07)."
      : "The context was compacted or resumed. No progress file exists for this branch. If the task is longer than one session, " +
        "create openspec/changes/<change-id>/progress.md from openspec/templates/progress.md (AI standard 01).";
  process.stdout.write(JSON.stringify({ hookSpecificOutput: { hookEventName: "SessionStart", additionalContext: text } }));
  process.exit(0);
}

checkSessionStart();
checkSecretRead();
checkShared();
checkPlan();
checkPush();
checkSubagentMerge();
checkCommit();
checkAfterEdit();
