---
name: to-tickets
description: Slice an OpenSpec change of this project into tasks and prepare or publish linked GitHub board tickets.
---

# Turn a change into tasks

Resolve the repository root. Read `openspec/changes/<change-id>/proposal.md`,
its design and deltas, and `docs/dev/ai-standard/02-harness.md` for board rules.
The board is a GitHub Project. Set `<owner>` and `<project-number>` below to your
board before you publish. If they are not set, stay in dry-run mode and report it.

Write `tasks.md` using `openspec/templates/tasks.md`. Make each slice an observable
outcome with named scenarios, checks, dependencies, Area, Slice number, and Size.
Preserve existing issue URLs and completed checkboxes. Every scenario must appear
in at least one slice. A draft proposal can be sliced but cannot be published.

## Prepare locally

Default to dry-run. Prepare one Markdown body per slice under
`openspec/changes/<change-id>/tickets/`. Include the change ID, proposal link,
scenario names, acceptance checks, dependencies, Area, Slice, and Size.
Use a stable marker `<!-- openspec:<change-id>:slice:<number> -->` in each body.
Report the proposed titles and files without calling a mutation command.
If the user forbids new issues, remain in this mode and reuse existing issues where applicable.

## Publish when authorized

Proceed only when the owner approved the proposal and the session authorizes
issue creation. A no-new-issues instruction overrides this workflow.

1. Determine the GitHub repository from `gh repo view`. Read existing issues,
   including closed issues, and match each stable marker or recorded issue URL.
   Reuse matching issues. If a matching issue is closed, report it instead of duplicating it.
2. Verify that the proposal is on a remote branch, so its GitHub link resolves.
   Inspect `gh project field-list <project-number> --owner <owner> --format json` and obtain
   current field and option IDs. Read board items before adding them.
3. For each missing issue, run `gh issue create --repo <repo> --title <title>
   --body-file <body-file> --label <appropriate-existing-label>`. Record its URL
   in `tasks.md` immediately. If a request times out, query by marker before retrying.
4. Add each issue to project `<project-number>` with `gh project item-add` if it is not already there.
   Set Status to Ready and Area, Slice, Size with `gh project item-edit`, using
   the returned item ID and discovered field IDs.
5. Read back the issues and board fields. Report created, reused, and incomplete items.
   A partial failure stops publication. Preserve URLs so the next run can resume.

Completion means that all requested slices have local bodies in dry-run mode,
or verified issue URLs and board fields in publish mode. Do not claim dry-run
output proves live issue creation or board updates.
