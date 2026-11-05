// Regression tests for guard.mjs. Run: task test:hooks
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { randomBytes } from "node:crypto";
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, realpathSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { after, before, test } from "node:test";

const guard = path.join(import.meta.dirname, "guard.mjs");
const project = path.join(import.meta.dirname, "..", "..");
let repo;

function sh(args, cwd = repo) {
  execFileSync("git", args, { cwd, stdio: "ignore" });
}

// A small repository on the branch chore/x, with a second worktree on main.
// It has a Go module for go vet, and the project mise.toml, so the hook
// finds the pinned tools.
before(() => {
  repo = realpathSync(mkdtempSync(path.join(tmpdir(), "guard-")));
  sh(["init", "-q", "-b", "main"]);
  mkdirSync(path.join(repo, "core/cmd/worker"), { recursive: true });
  writeFileSync(path.join(repo, "core/go.mod"), "module example.com/fixture\n\ngo 1.22\n");
  writeFileSync(path.join(repo, "core/cmd/worker/main.go"), "package main\n\nfunc main() {}\n");
  if (existsSync(path.join(project, "mise.toml"))) copyFileSync(path.join(project, "mise.toml"), path.join(repo, "mise.toml"));
  sh(["add", "-A"]);
  sh(["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init"]);
  sh(["checkout", "-q", "-b", "chore/x"]);
  sh(["worktree", "add", "-q", path.join(repo, "wt-main"), "main"]);
});

after(() => rmSync(repo, { recursive: true, force: true }));

function run(event) {
  const r = spawnSync("node", [guard], {
    input: JSON.stringify({ cwd: repo, ...event }),
    env: { ...process.env, CLAUDE_PROJECT_DIR: repo },
    encoding: "utf8",
    // go vet compiles the package on the first run.
    timeout: 30_000,
  });
  return r.status;
}

const bash = (command) => run({ hook_event_name: "PreToolUse", tool_name: "Bash", tool_input: { command } });
const subagentBash = (command) =>
  run({ hook_event_name: "PreToolUse", tool_name: "Bash", tool_input: { command }, agent_id: "a1", agent_type: "general-purpose" });
const edit = (file, extra = {}) =>
  run({ hook_event_name: "PreToolUse", tool_name: "Edit", tool_input: { file_path: path.join(repo, file) }, ...extra });

test("push to main is blocked", () => {
  for (const cmd of [
    "git push origin main",
    "git push origin 'HEAD:main'",
    "git push origin +main",
    "git push origin refs/heads/main",
    "git push origin :main",
    "git push origin --all",
    "git push --mirror origin",
    "git push --repo=origin HEAD:main",
    "git push origin 'refs/heads/*:refs/heads/*'",
    "git -C . push origin main",
    "git add -A && git push origin main",
    "git status | git push origin main",
    "git checkout main && git push origin",
    "git -C ../other push origin",
    "git push",
    "git push origin HEAD",
    "GIT_TRACE=1 git push origin main",
    "(cd x; git push origin main)",
    'branch=main; git push origin "$branch"',
    "git push origin `echo main`",
    "env GIT_TRACE=1 git push origin main",
    "command git push origin main",
    "sudo -E git push origin main",
    "sudo -u root git push origin main",
    "env -u UNUSED git push origin main",
    "env -u UNUSED sudo -u root git push origin main",
    "nice -n 5 git push origin main",
    "bash -c 'git push origin main'",
    "git push origin > /tmp/push.log",
    "bash -lc 'git push origin main'",
    "sh -e -c 'git push origin main'",
    "bash -c -e 'git push origin main'",
    "git push origin --signed main chore/x",
    "git push origin 2>&1 >/tmp/log",
  ]) {
    assert.equal(bash(cmd), 2, cmd);
  }
});

test("other pushes and commands pass", () => {
  for (const cmd of [
    "git push -u origin chore/x",
    "git push -u origin chore/x feat/y",
    "git push --repo=origin chore/x",
    "git push origin feature/main-page",
    "sudo -u root git push origin chore/x",
    "env -u UNUSED git push origin chore/x",
    "env -u UNUSED sudo -u root git push origin chore/x",
    "nice -n 5 git push origin chore/x",
    'git commit -m "push the fix to main later"',
    'git commit -m "document command; git push origin main"',
    "echo git push origin main",
    "git push -u origin chore/x > /tmp/push.log 2>&1",
    "git status",
  ]) {
    assert.equal(bash(cmd), 0, cmd);
  }
});

test("a subagent cannot merge a PR, the main agent can", () => {
  for (const cmd of [
    "gh pr merge 12 --squash",
    "gh pr merge --auto --squash 12",
    "gh api -X PUT repos/o/r/pulls/12/merge",
    "gh api repos/o/r/merges -f base=main -f head=chore/x",
    "bash -c 'gh pr merge 12'",
    "env GH_TOKEN=x gh pr merge 12",
  ]) {
    assert.equal(subagentBash(cmd), 2, cmd);
    assert.equal(bash(cmd), 0, `main agent: ${cmd}`);
  }
  for (const cmd of ["gh pr view 12", "gh pr checks 12", "gh api repos/o/r/pulls/12", "echo gh pr merge 12"]) {
    assert.equal(subagentBash(cmd), 0, cmd);
  }
});

test("a subagent cannot change shared files", () => {
  const sub = { agent_id: "a1", agent_type: "general-purpose" };
  for (const file of ["api/openapi.yaml", "core/migrations/00002_x.sql", "docs/dev/CONTEXT.md", "openspec/specs/x/spec.md"]) {
    assert.equal(edit(file, sub), 2, file);
  }
  assert.equal(edit("api/openapi.yaml"), 0, "the main agent can");
  assert.equal(edit("core/cmd/worker/main.go", sub), 0, "module code on a chore branch");
});

test("code on main is blocked, also in another worktree", () => {
  assert.equal(edit("wt-main/core/cmd/worker/main.go"), 2);
});

test("a feat branch needs an approved proposal", () => {
  sh(["checkout", "-q", "-b", "feat/try"]);
  try {
    assert.equal(edit("core/cmd/worker/main.go"), 2, "no proposal");
    const dir = path.join(repo, "openspec/changes/try");
    mkdirSync(dir, { recursive: true });
    writeFileSync(path.join(dir, "proposal.md"), "# Try\n\nStatus: draft\n");
    assert.equal(edit("core/cmd/worker/main.go"), 2, "draft proposal");
    writeFileSync(path.join(dir, "proposal.md"), "# Try\n\nStatus: approved\n");
    assert.equal(edit("core/cmd/worker/main.go"), 0, "approved proposal");
  } finally {
    sh(["checkout", "-q", "chore/x"]);
  }
});

// Files for the fast checks. They are written into the fixture by each test
// and removed after it.
const UNFORMATTED = "package main\n\nfunc   unformatted() {}\n";
const VET_ERROR = 'package main\n\nimport "fmt"\n\nfunc wrongFormat() {\n\tfmt.Printf("%d\\n", "s")\n}\n';

function write(file, text) {
  writeFileSync(path.join(repo, file), text);
}

// Back to the clean commit. git clean does not remove the wt-main worktree.
function reset() {
  sh(["reset", "-q", "--hard"]);
  sh(["clean", "-qfd"]);
}

const postEdit = (file) =>
  run({ hook_event_name: "PostToolUse", tool_name: "Edit", tool_input: { file_path: path.join(repo, file) } });

test("fast checks after an edit report problems", () => {
  try {
    write("core/cmd/worker/bad.go", UNFORMATTED);
    assert.equal(postEdit("core/cmd/worker/bad.go"), 2, "unformatted");
    write("core/cmd/worker/bad.go", VET_ERROR);
    assert.equal(postEdit("core/cmd/worker/bad.go"), 2, "go vet error");
    assert.equal(postEdit("core/cmd/worker/main.go"), 2, "same package as the vet error");
    rmSync(path.join(repo, "core/cmd/worker/bad.go"));
    assert.equal(postEdit("core/cmd/worker/main.go"), 0, "clean file");
  } finally {
    reset();
  }
});

test("a commit with failing fast checks is blocked", () => {
  try {
    write("core/cmd/worker/bad.go", UNFORMATTED);
    sh(["add", "core/cmd/worker/bad.go"]);
    for (const cmd of [
      "git commit -m x",
      "git add -A && git commit -m x",
      "bash -c 'git commit -m x'",
      "env A=1 git commit -m x",
      `git -C ${repo} commit -m x`,
      "git -c user.name=t commit --amend --no-edit",
    ]) {
      assert.equal(bash(cmd), 2, cmd);
    }
    reset();
    // A changed tracked file: git commit -a would include it.
    write("core/cmd/worker/main.go", "package main\n\nfunc   main() {}\n");
    assert.equal(bash("git commit -am x"), 2, "changed, not staged");
    reset();
    // A new file: git add -A runs after the hook.
    write("core/cmd/worker/bad.go", UNFORMATTED);
    assert.equal(bash("git add -A && git commit -m x"), 2, "new, not added yet");
    reset();
    write("core/cmd/worker/vet.go", VET_ERROR);
    sh(["add", "-A"]);
    assert.equal(bash("git commit -m x"), 2, "go vet error");
  } finally {
    reset();
  }
});

// oxlint and oxfmt come from the web/node_modules of the project that holds
// this hook, so the fixture needs no install. Skip where they are missing.
const hasWebTools = () => existsSync(path.join(project, "web", "node_modules", ".bin", "oxfmt"));

test("fast checks report unformatted TypeScript", (t) => {
  if (!hasWebTools()) return t.skip("web/node_modules is not installed");
  mkdirSync(path.join(repo, "web/src"), { recursive: true });
  try {
    write("web/src/x.ts", "export const   x =  1\n");
    assert.equal(postEdit("web/src/x.ts"), 2, "after an edit");
    assert.equal(bash("git add -A && git commit -m x"), 2, "before a commit");
    write("web/src/x.ts", "export const x = 1;\n");
    assert.equal(postEdit("web/src/x.ts"), 0, "formatted file after an edit");
    assert.equal(bash("git add -A && git commit -m x"), 0, "formatted file before a commit");
  } finally {
    reset();
  }
});

// gitleaks comes from mise.toml. Skip the test where it is not installed.
const hasGitleaks = () =>
  spawnSync("mise", ["which", "-C", repo, "gitleaks"]).status === 0 || spawnSync("gitleaks", ["version"]).status === 0;

test("a commit with a secret is blocked", (t) => {
  if (!hasGitleaks()) return t.skip("gitleaks is not installed");
  // A fake GitHub token, made at run time, so the repository has no token.
  const token = "ghp_" + randomBytes(48).toString("base64").replace(/[^A-Za-z0-9]/g, "").slice(0, 36);
  try {
    write("settings.env", `GITHUB_TOKEN=${token}\n`);
    assert.equal(bash("git add -A && git commit -m x"), 2, "new file");
    sh(["add", "settings.env"]);
    assert.equal(bash("git commit -m x"), 2, "staged file");
  } finally {
    reset();
  }
});

test("a commit with clean files passes", () => {
  write("core/cmd/worker/ok.go", "package main\n\nfunc ok() {}\n");
  try {
    for (const cmd of [
      "git commit -m x",
      "git add -A && git commit -m x",
      "echo git commit -m x",
      "git log --grep commit",
      "git commit-tree HEAD^{tree} -m x",
    ]) {
      assert.equal(bash(cmd), 0, cmd);
    }
  } finally {
    reset();
  }
});
