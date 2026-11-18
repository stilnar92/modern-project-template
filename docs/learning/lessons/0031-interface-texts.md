# 0031 — Interface texts in more than one language

From PR #111 · 12.11.2026 · Topics: i18n, bundle size, type safety

## Problem

The interface texts were literals in the components. The product starts in Russian, and other languages can come later. Two apps show some of the same texts. One of them, the site, is static pages that send no JavaScript. We need one place for the texts, and a way to find a text that is missing.

## Predict first

A component calls `t("services_title")`, but the key in the dictionary is `service_title`. The dictionary is a JSON file that loads at run time. When does someone see the mistake, and what is on the screen until then?

## Theory

An i18n system has three parts: where the texts live, how the code refers to a text, and when the lookup happens.

**Run-time lookup.** The classic model sends a dictionary and a library to the browser. The code calls `t("key")`. The key is a string, so the compiler cannot check it. A missing key shows the key or an empty place, and only a person or a test sees it. The bundler cannot know which keys the code uses, so the full dictionary goes to the browser.

**Compiled messages.** A compiler reads the message files before the build and writes one function for each message. The code calls `m.service_title()`. Now a message is a normal export:

- A key that does not exist is an import that does not exist. The type check fails.
- A message with a parameter is a function with a typed argument.
- The bundler removes each function that no code imports (tree shaking). An unused message does not reach the browser.
- A page that renders on the server or at build time calls the function there. It sends only the text.

**Plurals** are not "one or many". The Unicode CLDR defines categories for each language: `zero`, `one`, `two`, `few`, `many`, `other`. English uses two of them. Russian uses `one` (1, 21), `few` (2 to 4, 22 to 24) and `many` (5 to 20, 11, 12). The browser has these rules in `Intl.PluralRules`, so a library does not need its own tables.

Sources:

- [Paraglide JS](https://inlang.com/m/gerre34r/library-inlang-paraglideJs), inlang: the compiler model and the message functions.
- [MDN: `Intl.PluralRules`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/PluralRules): the plural categories in the browser.
- [Unicode CLDR: Plural Rules](https://cldr.unicode.org/index/cldr-spec/plural-rules): the categories and why they differ.
- [Reduce JavaScript payloads with tree shaking](https://web.dev/articles/reduce-javascript-payloads-with-tree-shaking), web.dev: how a bundler removes unused exports.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| i18next | A run-time library and dictionaries with string keys | A run time in the browser, also on the site. The compiler does not check the keys |
| Lingui | Messages are marked in the code, extracted to catalogs and compiled | Not taken. The record gives no reason |
| Our own code | Typed modules and `Intl` | We write and maintain the compiler and the plural handling |
| Paraglide | A compiler writes one typed function for each message | A generation step before each build and type check |

## Edge cases

| What if | What happens |
| --- | --- |
| The code uses a key that is not in the file | The function does not exist. The type check fails |
| A message is in the file, and no code uses it | The bundler removes its function. It costs nothing in the browser |
| A count needs a Russian plural | The message has the `plural` selector with `one`, `few` and `many`. A test checks 1, 21, 2, 24, 5, 11, 12 and 100 |
| The site shows a text | The page calls the function at build time. The production build still has 0 JavaScript files |
| The compiler plugin loads from a local path, so CI does not need the CDN | The compiler wrote zero messages and gave no error. The tests found it. The plugin loads from the CDN with a pinned version |
| The generated folder goes into git | The compiler writes a README with an absolute path. So the output is not committed. `postinstall` and `mise run gen` write it |
| A developer installs a fresh copy | `postinstall` compiles the messages. The imports work before the first build |

## Decision

We use Paraglide in one shared package. Both apps import `m` from it. Russian is the base locale and the only locale now. The reasons: compiled typed functions, no run time on the site, and Russian plurals through `Intl.PluralRules`.

The compiled output is a build product, not a source. It is not in git, and the linter and the formatter skip it.

This would be wrong if people who do not write code must change the texts without a build. A compiled message needs a build to change. It also depends on a CDN at generation time. If that becomes a problem for CI, the plugin loading needs a new solution.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), "i18n".

## In our code

1. `web/packages/i18n/messages/ru.json`: the texts. See `photos_count` for a plural.
2. `web/packages/i18n/project.inlang/settings.json`: the locales and the pinned plugin.
3. `web/packages/i18n/src/index.ts` and `messages.test.ts`: the exports and the plural proof.
4. `web/apps/dashboard/src/errors/messages.ts`: message functions as values, one pair for each error kind.

## Check yourself

1. Why can a bundler remove an unused message with compiled functions, and not with a `t("key")` dictionary?
2. A message for a count has only two forms, "one" and "other". What does a Russian user see for 3 and for 5?
3. The compiler finished with no error, and the app showed no texts. What does this say about the checks that a generation step needs?

<details><summary>Answers</summary>

1. A compiled message is an export. The bundler sees each import and removes the exports that nobody imports. A dictionary is one object, and the key is a string that can come from a variable. The bundler cannot prove that a key is unused, so it keeps all of them.
2. Both numbers get the "other" form. Russian needs "фотографии" for 3 (`few`) and "фотографий" for 5 (`many`). One form cannot be correct for both, so one of the two phrases is wrong.
3. A generator can succeed and write nothing. An exit code of zero is not a proof. A test must call a real message and compare its text. Here such tests failed and showed the problem.

</details>
