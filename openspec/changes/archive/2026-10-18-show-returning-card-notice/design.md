# Design

## Context

- `StudyScreen._rate` (`lib/features/study/study_screen.dart`) answers through `golang.study.answer` and shows the returned `StudyState`; `_apply` treats a card with another ID as a new card, but the answered card coming back has the same ID, and `_apply` still resets it to its question.
- Go picks the next card in `core` (`golang/core/queue.go`): with nothing else due, a learning card due within the learn ahead limit (`Settings.LearnAhead`) is shown before its due time. `StudyCard` (`golang/svc_study.go`, built in `studyCard`) does not say whether a card is shown early, or when it is due.
- The study screen already shows an inline notice for empty cards (`_EmptyCardNotice`), between the top bar and the card, which neither covers the card nor the buttons. Snack bars are avoided during study because they cover the buttons.
- `formatSpent` (`lib/widgets/time_format.dart`) formats durations as "11 minutes" or "1 h 5 min".

## Goals / Non-Goals

**Goals:**
- Make it obvious that the click registered when the answered card comes straight back, or another card that looks alike comes (Image Occlusion notes of one image).
- Explain why a card is shown before it is due while the counts show nothing due.

**Non-Goals:**
- Changing which card comes next, or the learn ahead limit.

## Decisions

### 1. Go reports the early due time

`StudyCard` gains `DueEarly *time.Time` (`json:"dueEarly,omitempty"`), set in `studyCard` when the card is a learning card whose due time is after now, so it is shown because of learn ahead. Go decides, because it knows the card's due time and the clock it scheduled with; the app only formats it. A field rather than a flag keeps the time with the reason.

Alternative: use the label of the pressed button's interval preview ("<11m"). Rejected: the label is rounded and estimated before the answer, and it does not say whether the card is shown early.

### 2. The app detects the returning card

After `_rate`, when the new state's card has the ID of the card that was answered, the screen remembers that the notice says "this card again" for this showing (`_shownAgain`). The flag is set only by `_rate`, so undo, a re-render (`_renderAgain`) and a learning card becoming due do not set it. It is cleared when the answer is revealed and when another card is applied.

Whether a card is shown early needs no flag: it is `dueEarly` of the card on screen. When the card becomes due while its question is shown, the screen fetches the state again (`checkAt`) and the early notice goes, or says "due again now" for the card just answered.

### 3. Inline notice, like the empty-card notice

`_ReturningCardNotice` sits where `_EmptyCardNotice` does, in the same style, on the question only. Its text:

- card shown early: "Shown early: due in 11 minutes, because nothing else is due."
- the card just answered, shown early: "This card again: due in 11 minutes, shown early because nothing else is due."
- the card just answered, already due: "This card again: it is due again now."

The time is from `DueEarly`: `formatSpent` of the time left, rounded up to whole minutes, at least one minute. It is announced to screen readers as a live region, since the visual change is otherwise small.

## Risks / Trade-offs

- [The notice is one more thing on screen] → It shows only on early or returning cards, and goes away with the answer. Users who prefer to wait set the learn ahead limit to 0.
- [Auto advance shows the question again without a click] → The notice still applies: the card was answered and came back, which the notice explains.
