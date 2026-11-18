# UX standard

**Status:** ACCEPTED ([ADR 0011](../adr/0011-ux-standard.md))

These are the UX rules for every change that a person sees: the dashboard, the public pages, the widget, bot messages and the survey. Agents and people apply them when they design, write and review the code. They are not a skill to call. They are rules, the same as the [frontend standard](../frontend-standard/README.md) and the [backend standard](../backend-standard/README.md).

The source is the `ux-best-practices` skill of [CodeAlive-AI/ai-driven-development](https://github.com/CodeAlive-AI/ai-driven-development/tree/main/skills/ux-best-practices) (MIT, commit `bcc63f2`, sources verified 2026-10-21). We adapted its workflow into the rules below. The [reference](#reference) folder has its topical files without changes: the exact criteria, the sources and the templates.

## Who our users are

- **Masters** use the dashboard. They often use a phone, between clients, with one hand.
- **Clients of masters** open the public pages and the widget from a link in a messenger or a social network. Most of them use a phone.
- The interface language is the language of the user (Russian first). The rules for the text are in [U31](#text).

Every rule below applies to the narrow phone screen first.

## Baseline

- **U1.** The accessibility target is **WCAG 2.2 AA**. Apply AAA only as a labeled improvement. This target is not a legal claim.
- **U2.** Give each size with its unit. On the web, use CSS px. Do not mix CSS px, iOS pt and Android dp.

## Design the flow before the screen

- **U3. Start from the task.** For each new screen or flow, write in the spec: the user task, the user, the device, the cost of an error and the observable success. Do not invent personas, user quotes or metrics.
- **U4. Walk each step.** For each step of the flow, check five things:
  1. **Discovery:** the person sees where to start and which option fits the task.
  2. **Understanding:** the person knows what is necessary, why, and what happens next.
  3. **Action:** the person can do it with touch and with a keyboard, without memorizing data from another screen.
  4. **Feedback:** waiting, saving, success and failure look different.
  5. **Recovery:** the person can correct input, go back, undo or retry safely.
- **U5. List the states.** The spec lists the states that the flow has, and the code shows each of them: initial, loading, with data, empty, no results, validation error, service error, no permission, success. Add offline, stale data, partial result, unknown result and confirmation of a destructive action when the flow has them. Do not add states that the flow does not have.
- **U6. Keep the context.** Input, filters, the selected item and the focus stay after a transition, an error and a return. A failed submit never clears the form.
- **U7. Keep familiar patterns.** Use the design system (Mantine in the dashboard) and the patterns that people know from other apps. A new pattern needs a reason in the spec.

## Forms

- **U8. Ask only for necessary data.** Explain why a field that is not usual is necessary.
- **U9. Label each field.** Use `label`, and `fieldset` with `legend` for a group. A placeholder is not a label.
- **U10. Validate on submit.** Do not show an error for input that the person has not finished. Show an early hint only when it helps, for example a length limit. The backend validates again and owns the rule.
- **U11. Accept clear variants.** Accept spaces, dashes and paste in a phone number or a code. Do not guess a date, a currency or an identifier when the input is not clear. Do not remove meaningful characters from names.
- **U12. Help the browser fill the form.** Set the correct `type`, `inputmode` and `autocomplete`. Allow paste. Password managers must work (WCAG 3.3.8). Do not ask again for data that the person gave earlier in the same process (WCAG 3.3.7).
- **U13. Show errors where they are.** Put the message next to the field. For several errors, also show a summary with links to the fields, and move the focus to it. The message says what is wrong and how to correct it. Do not blame the person.
- **U14. Name the error type correctly.** An input error, a refusal by a business rule, a missing permission and a service failure are different messages. Do not ask the person to correct a field when the cause is not in the field.
- **U15. Name a button after its result.** Write "Отправить заявку" or "Сохранить изменения", not "OK" or "Далее" when the result is known.
- **U16. Review before an important submit.** Before a booking, a payment or another hard-to-undo action, show the data, the full known price and the terms. Let the person change one part without starting again.

## Feedback and waiting

- **U17. Respond at once.** Each action gets a visible response. Do not show success before the server confirms it.
- **U18. Show loading at the scale of the area.** Use a local indicator for one action and a skeleton for a known structure. Keep other areas usable. Do not show a progress percentage that you cannot measure.
- **U19. Reserve the space.** Content that loads must not move the page (CLS 0.1 or less, [frontend standard 30](../frontend-standard/30-performance.md)).
- **U20. Check an unknown result before a retry.** When the response to a submit is lost, find the state of the operation first. Never offer a blind retry of a booking or a payment. This rule and the double-tap guard in [web/CLAUDE.md](../../../web/CLAUDE.md) prevent duplicates.
- **U21. Protect destructive actions.** Give undo, or a confirmation that names the consequence. Do not put "Delete" next to a frequent action.

## Keyboard, focus and semantics

- **U22. Use native elements first.** A custom control needs its own keyboard handling, name, role and state. `role="button"` does not add key handling.
- **U23. Make every action work with a keyboard.** The focus is always visible. A sticky panel or a banner does not fully hide the focused element (WCAG 2.4.11).
- **U24. Use a modal dialog only for a separate decision.** The focus moves into the dialog and stays there. Escape and a visible close button close it. The focus then goes back to the element that opened it. A Mantine component does not prove that our integration is correct: check it.
- **U25. Announce status without moving focus.** Use `role="status"` or `aria-live` for "Saved" and similar messages (WCAG 4.1.3).
- **U26. Do not use color as the only signal.** An icon-only button has an accessible name. An icon that is not clear has a text label.

## Size, contrast and layout

- **U27. Make targets large enough.** The minimum is 24 × 24 CSS px or a WCAG 2.5.8 exception. For the main actions on a touch screen, use 44 × 44 CSS px. This is our choice for phone users, not an AA requirement.
- **U28. Keep the contrast.** Text has 4.5:1, large text 3:1. Controls, their states and meaningful icons have 3:1 against the adjacent colors. Measure the colors from the tokens, not from a screenshot.
- **U29. Support reflow and zoom.** At 320 CSS px width and at 200% text size, all content and actions are available without horizontal scroll. This adds to D8 and D14 in [frontend standard 18](../frontend-standard/18-design-principles.md).
- **U30. Do not hide important actions behind hover.** Touch screens have no hover.

## Text

- **U31. Plain language.** Use the words of the user, short sentences and one name for one thing. Information, success, warning and error look and read differently. Put a message at the level of the area that it is about. Do not make every message a system banner.
- **U32. Explain an empty state.** First use, no results and a load error are different screens. Each one says why it is empty and gives the next useful step. A filter with no results offers to change the filter, not to create data.

## Honesty

- **U33. No deceptive patterns.** No hidden costs, fake urgency, fake reviews, forced consent, preselected paid options or hard cancellation. Show the full known price and the terms before the final action.
- **U34. Be honest about AI.** When a feature uses AI, show the difference between a suggestion, an action in progress and a confirmed result. Do not invent confidence levels or sources.

## How to use the laws of UX

- **U35.** Use the [laws of UX](reference/cognitive-principles.md) to explain a problem and to make a hypothesis. They are not numeric limits. Miller does not mean "seven menu items". Hick does not forbid a long catalog. Fitts is not a reason to make every button larger. A beautiful screen is not proof of usability.

## Verification of a UI change

A UI change is done only when you checked it in a browser. Use Playwright or Chrome DevTools.

1. Open the changed path and each state from U5 that it has.
2. Do the task with a keyboard only.
3. Check the widths 320 and 1280 CSS px, and 200% text size. For a key page, also run the real viewports in `web/e2e/tests/viewports.spec.ts`.
4. Slow down the network and make the request fail once.
5. Run an automated accessibility scan if one is available. A scan finds only a part of the problems.

In the PR, write what you checked and what you did not check. Use the words "not verified" for a check that you did not do. A screenshot cannot prove keyboard behavior, contrast values, screen reader output or speed. A walk-through by an AI agent is not research with real users.

## UX review and audit

When you review a screen or the owner asks for a UX audit:

- Write each finding as: observation → affected task → likely consequence → change → basis → verification.
- Give the exact change: the button text, the step order or the behavior. Do not write "make it more intuitive".
- Mark each finding **confirmed**, **likely risk** or **not verified**.
- Give the severity on the S0–S3 scale. S3: a key task is impossible, or data loss is possible. S0: polish. Keep severity apart from reach and cost.
- Give the basis: a WCAG criterion with its number and level, or a source ID from [sources.md](reference/sources.md). Label your own conclusion as your own.

The templates are in [output-templates.md](reference/output-templates.md).

## Reference

The files in [reference/](reference/) are a copy of the source without changes. Read only the file that the task needs.

| Task | File |
| --- | --- |
| Hierarchy, cognitive load, navigation, onboarding, laws of UX | [cognitive-principles.md](reference/cognitive-principles.md) |
| Forms, errors, states, modals, tables, responsive layout, text | [interaction-patterns.md](reference/interaction-patterns.md) |
| Keyboard, semantics, target size, contrast, zoom, sign-in | [accessibility.md](reference/accessibility.md) |
| AI features: trust, control, recovery | [ai-ux.md](reference/ai-ux.md) |
| Catalog, search, checkout, price and terms | [commerce.md](reference/commerce.md) |
| Loading, response time, layout stability | [performance.md](reference/performance.md) |
| Research method, metrics, how to verify a finding | [research-and-validation.md](reference/research-and-validation.md) |
| Audit, flow spec and validation plan templates | [output-templates.md](reference/output-templates.md) |
| Source, version and date of a claim | [sources.md](reference/sources.md) |

When a rule above and a reference file disagree, the rule above wins: it has our product context. We did not copy the `audit/` folder of the source (about 5,000 lines). Its topical files above have the exact criteria. To update the copy, compare it with the source and record the new commit in this file.
