# Design

## Context

- Every mutation in `golang/core` builds one `op` (changed and deleted decks, note types, notes, cards, settings, a review entry or its removal) and calls `commit`, which journals it before applying it in memory.
- Study undo uses the review entry's `Before` state and `UndoReview`.
- Compaction deletes media no note references.

## Goals / Non-Goals

**Goals:** generic undo for every op without per-feature code; durable and crash-safe; bounded memory.

**Non-Goals:** persisting the undo history across restarts; undoing restore or "start fresh".

## Decisions

### 1. Inverse ops from before-images

Before `commit` applies an op, it captures the current versions of everything the op touches (or their absence). The inverse op re-puts the old versions and deletes what the op created. Answers already have an exact inverse (`ReviewUndone` plus the `Before` card state). Undo commits the inverse as a normal op, so it is journaled and announced; redo commits the original again (with refreshed before-images).

Alternative: snapshots per step. Rejected: too large for 50,000-card collections.

### 2. Bounds

Keep 30 steps or 50 MB of before-images, whichever is reached first; older steps are dropped. A single op larger than the limit (a huge import) is still undoable as the most recent step, but drops all older steps.

### 3. Media

Compaction treats media referenced by notes in the undo history as referenced, so undoing a note delete brings back working images.

### 4. Interaction with study

The study session's undo stack stays a filtered view of the history (answers of this session). Undoing in the browser while a session is open reloads the session state through the existing change notifications.

## Risks / Trade-offs

- [Missed before-image for a new op kind] → The capture is generic in `commit`, not per mutation; a test iterates every exported mutation and checks undo restores the snapshot byte for byte.
- [Undo after the user changed the same objects elsewhere] → Undo is linear: it reverts the latest step only, so this cannot happen within one app session.
