# Design

## Context

- Manual buries are stored on the card (`Card.BuriedUntil`). Sibling burying is not stored: `Collection.view` excludes new and review cards of notes answered today.
- `NextDue(deckID)` already computes when something becomes due.

## Goals / Non-Goals

**Goals:** a cheap overview call; unbury that covers both kinds of burying.

**Non-Goals:** a rich text editor for descriptions (a plain multi-line field with HTML allowed is enough), markdown.

## Decisions

### 1. Buried count

Count cards in the subtree that are manually buried (`BuriedUntil` after now) plus cards excluded today by sibling burying that would otherwise be due. Computed in the same pass as the queue (`dayView` already knows both).

### 2. Unbury

Clear `BuriedUntil` on manually buried cards in the subtree, and store an "unburied" marker on each deck of the subtree: `UnburiedAt time.Time`, the time of the unbury. Sibling burying in that deck then only counts answers given after it, so the cards held back so far come back, while notes studied after the unbury still bury their siblings as usual (as in Anki). Only today's answers bury siblings, so the marker expires by itself the next day and no cleanup is needed.

(Planned first as `UnburiedDay int`, which would have turned sibling burying off for the rest of the day; changed during implementation.)

Alternative: store sibling buries explicitly like manual buries. Rejected: it would add a write per answer and change today's model.

### 3. Description rendering

Shown with the same HTML widget as cards, sanitised in Go the same way (no scripts), with links opened externally.

## Risks / Trade-offs

- [An extra tap before studying] → The deck list keeps a direct Study action in each deck's menu.
