# Design

## Context

`answer` in `golang/render/render.go` renders the question with `Question` and stores it as `{{FrontSide}}` after removing the input placeholder `<cf-typeans></cf-typeans>`. Only `typeAnswer` running in the answer template's own context (`golang/render/template.go`) produces a comparison. So a `type:` filter that reaches the answer only through `{{FrontSide}}` disappears.

In Anki, both sides render `{{type:…}}` to a `[[type:…]]` marker. The reviewer then replaces every marker in the answer HTML, including the one copied in by `{{FrontSide}}`. When it does, it also removes `<hr id=answer>` and puts it in front of the comparison, so that the comparison appears below the line.

## Goals / Non-Goals

**Goals:**
- Show the comparison for question-only `type:` filters, at the place they hold in `{{FrontSide}}`, with their modifiers.
- Keep templates that already work rendering the same output.

**Non-Goals:**
- Copying Anki's double comparison when both sides have a `type:` filter and the answer uses `{{FrontSide}}`.
- Changing the study screen, `StudyService.Reveal` or the input on the question.

## Decisions

**Render a marker into `{{FrontSide}}` and resolve it after the answer is rendered.** Rendering the question records the first active `type:` filter: its field value after `cloze:` and its `nc` flag. `{{FrontSide}}` keeps a private marker where the input placeholder was. After the answer template is rendered:
- if its context rendered a `type:` filter of its own (`typeUsed`), the marker is removed, as the placeholder is today;
- otherwise the first marker becomes the recorded filter's comparison, or its expected text when nothing was typed. Any later markers (from `{{FrontSide}}` used twice) are removed.

The marker must survive template substitution but not field content. A string that sanitizing would never let through from a field, like the existing placeholder, is enough.
*Alternative:* render the answer template first to find out whether it has a `type:` filter, then build `{{FrontSide}}`. This renders the answer twice and still needs the question's filter details, so it is no simpler.

**Show the comparison once.** When the answer has its own `type:` filter, `{{FrontSide}}` shows nothing in its place. This is today's tested behavior (`TestRenderCardTyped`, `TestTypeAnswer`) for the most common Anki type-in template. Anki would show the comparison twice there, which users see as a bug.

**Move `<hr id=answer>` as Anki does, only for the `{{FrontSide}}` case.** When the comparison comes from `{{FrontSide}}` and the rendered answer contains an `<hr id=answer>` line (matched by the existing `answerLine` pattern), the first such line is removed and placed in front of the comparison. `AnswerPart` then puts the comparison in the browser's answer column and in text exports, and the layout matches Anki. When the answer has no such line, nothing moves.

## Risks / Trade-offs

- [Question content after the `type:` filter now appears below the answer line once it moves] → This is how Anki shows the same template, and the reason for the move.
- [Templates that put `{{type:…}}` on the question just to get an input, and never wanted a comparison] → Anki shows the comparison for them too. Parity is the goal.
- [Marker leaking into output if a code path forgets to resolve it] → Resolve it in `answer`, the only place `{{FrontSide}}` is built, and test that no rendered answer contains it.
