# If the Ambiguity Won't Resolve Itself

Read this file when a naming question stays open after the lookup. It was moved out of `SKILL.md` to keep that file short.

When a name lands in `## Unresolved` — two spellings for one concept, one spelling for two
concepts, no obvious winner — the current tree can't settle it, but the repository's history
often can: which identifier was born first, which commit removed one while adding the other,
which one is dying.

**Offer it, don't run it silently.** Whenever you present `## Unresolved` items — after
generation, after an audit, or when a naming question hits one — offer once:

> "I can mine the git history for these — when each name was born, which replaced which,
> which is growing vs dying. Temporary index outside the repo, deleted afterwards.
> Want me to try that before you answer them by hand?"

Skip the offer if there is no `.git`, history is shorter than ~50 commits or squashed from
an import, or the items are `[WHITE-SPOT]` tags (an unnamed concept leaves no trace).

If the user accepts, read [references/git-history-mining.md](git-history-mining.md)
and follow it. The short version:

```bash
S=<skill-dir>/scripts/git_term_index.py
python3 $S build --repo-dir . --content   # throwaway SQLite index in $TMPDIR, never in the repo
python3 $S query Account Customer         # birth, dormancy, trajectory, renames, messages
python3 $S pair User Customer             # competing names: birth order + swap commits
python3 $S contexts Account               # where the name lives — the polysemy check
python3 $S search 'rename account'        # BM25 search over commit messages
python3 $S clean                          # delete the index when done
```

The index covers the **whole** diff history, so "born" means born. Use `--pathspec src/`
on large repos — it cuts build time and sharpens the signal at once.

`pair` ends in a labelled verdict — **RENAME (strong / probable / possible)**, **DRIFT**,
**NOT A RENAME**, **COEXISTENCE** — with the direction inferred from evidence, not argument
order. **Report the label as given; do not upgrade it.** "RENAME — strong" means commits
exchange the names *and* a subject announces it; everything weaker needs `git show` first.

For **polysemy** (one word, two meanings, two modules) use the path split: `pair` prints
`files: A in N, B in M, both in K`, and `contexts <name>` gives the directory breakdown for a
single word. `both in 0` — no file ever contained both — is real evidence for two bounded
contexts; shared files mean synonym drift, which history cannot settle. **Only claim a path
split when the tool printed one.** File-only renames (the file moved, the identifier did not)
appear in `query`'s file-renames section, not in `pair` — run both.

**History is evidence, not authority.** It ranks candidates and cites commits; the user
decides. Report proposals in one batch with confidence levels, apply only what is approved,
and record the commit behind each applied decision (`— renamed in a41f2c9` on the Legacy
line, or a `- **History**:` line on the entry).
