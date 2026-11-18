# 06 — Language and writing style

**Status:** DRAFT
**Decisions:** A14

## Which language where

All project text is in English and follows ASD-STE100.

| Text | Language | Style |
| --- | --- | --- |
| Development documents (`docs/dev/`) | English | ASD-STE100 |
| Issues, PR descriptions, PR comments | English | ASD-STE100 |
| Commit messages | English | ASD-STE100 |
| Code comments, logs, developer error messages | English | ASD-STE100 |
| Agent replies to the owner | English | ASD-STE100 |
| The frontend standard (`frontend-standard/`) | English | As-is: copied without changes ([ADR 0006](../adr/0006-ai-standard-project-wide.md)) |
| `docs/learning/`, `README.md` | English | ASD-STE100 |

**Exceptions:** text for product users (the interface, emails, notifications) is in the user's language. Quotes from user interviews stay in the original language.

## ASD-STE100

[ASD-STE100](https://www.asd-ste100.org/) is Simplified Technical English: a standard for simple technical English. English is not the owner's first language. STE makes text short, clear and easy to read.

| Rule | Example |
| --- | --- |
| A sentence in an instruction has a maximum of 20 words; in a description, a maximum of 25 | — |
| One instruction is one sentence, in the imperative | "Run `mise run test`." |
| Use the active voice | "The API returns 200", not "200 is returned by the API" |
| Use simple words with one meaning | "do", not "perform"; "start", not "initiate" |
| Do not drop articles | "Open the file", not "Open file" |
| No idioms, no slang | "It works", not "It's good to go" |
| A paragraph has one topic and a maximum of 6 sentences | — |
| Put a warning before the step that it applies to | "CAUTION: This deletes the data. Run `task db:reset`." |
| Use lists and tables instead of long paragraphs | — |

## Self-check before you publish text

Do this check for each text before you publish it. Source: the `ste-writing` skill of [Archiplan](https://github.com/archiplan-ai/Archiplan) (MIT), adapted.

1. Is a sentence longer than 20 words in an instruction, or 25 words in a description? Split it.
2. Is there a semicolon? Use two sentences.
3. Is there a contraction ("don't", "it's")? Write the full form.
4. Is there a passive sentence with a known actor? Make it active.
5. Is there a noun instead of a verb ("perform an analysis")? Use the verb ("analyze").
6. Is there a phrasal verb ("spin up", "figure out")? Use a plain verb ("start", "find").
7. Does one thing have two names? Select one name and use it everywhere.

**Plain words:**

| Write | Not |
| --- | --- |
| start | begin, commence, initiate |
| use | utilize, leverage |
| help | facilitate |
| make sure | ensure |
| before / after | prior to / subsequent to |
| about | regarding, concerning |
| get | obtain, acquire |
| show | demonstrate |
| also | additionally, furthermore, moreover |

No marketing words: seamless, robust, powerful, cutting-edge, effortless.

**Two modes:**

- **Strict:** procedures, runbooks, error messages, safety text. Apply every rule.
- **Normal:** documents, READMEs, PR descriptions. Apply the sentence, paragraph and verb rules, but you can use a wider vocabulary, so the text reads naturally.

STE fixes the form of a text. It does not make an empty text useful: the content must still be correct.

**Existing documents:** some English documents in `docs/dev` still contain semicolons. Fix them when you change the document.

## How it is checked

- The fresh-context review ([03](03-review-verification.md)) checks the text against the STE rules.
- The rule is in the root `CLAUDE.md`, section "Language", so the agent sees it in each session.
