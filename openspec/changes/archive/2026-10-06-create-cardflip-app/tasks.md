# Tasks

## 1. Project bootstrap

- [x] 1.1 Ensure `~/go/bin` is on `PATH` for the session. Then run `fluffy create --fvm 3.47.0 --patch-gtk --org dev.cardflip --project-name cardflip --platforms linux,windows,macos,android,ios,web .` at the repo root, falling back to `fvm use 3.47.0` + `fvm flutter create … .` + `fluffy init --patch-gtk` as described in design D1. Verify `.fvmrc`, `golang/`, `lib/golang/` and all six platform folders exist, and that `openspec/` and `.claude/` are untouched.
- [x] 1.2 Remove the fluffy demo code (demo exports in `golang/api.go`, demo UI in `lib/main.dart`), leaving a minimal `fluffy.Start`/`fluffy.Serve()` skeleton and a placeholder app. Verify `fluffy generate` runs cleanly and `fvm flutter analyze` reports no issues.
- [x] 1.3 Add the Go dependencies (`go-fsrs`, `sqlittle`, `klauspost/compress`, `google.golang.org/protobuf`, `tdewolff/parse/v2`) and the Dart dependencies (`path_provider`, `file_picker`, `flutter_widget_from_html_core`). Verify `go build ./...` and `GOOS=js GOARCH=wasm go build ./...` succeed in `golang/`, and `fvm flutter pub get` succeeds.
- [x] 1.4 Create `tool/check.sh`, which runs `go vet`, `GOOS=js GOARCH=wasm go vet`, `go test`, `fluffy generate --check`, `fvm flutter analyze` and `fvm flutter test`. Initialise git with a `.gitignore` that keeps the generated bindings committed. Verify `tool/check.sh` exits 0, and `fluffy run -d linux` shows the placeholder app.

## 2. Anki package reading spike and fixtures

- [x] 2.1 Write `tool/make_apkg_fixtures.py`, which uses the `anki` Python package to produce small fixture packages: a modern `anki21b` package, a legacy `anki2`/`anki21` package, packages with cloze, reversed, image media, audio media, suspended and review-state cards, and a truncated (corrupt) package. Commit them under `golang/importer/apkg/testdata/` with a README. Verify the script regenerates identical fixture contents.
- [x] 2.2 Write a spike test that opens each fixture and lists `notes` and `cards` row counts. It must cover zstd decompression of `collection.anki21b`, sqlittle table scans, and decoding of one `templates.config` protobuf. Verify it passes natively and builds under `GOOS=js GOARCH=wasm`. If sqlittle fails, replace it with a minimal read-only B-tree reader in `importer/apkg/sqlite` and record the decision in design.md.

## 3. Storage layer (persistence)

- [x] 3.1 Implement the `storage.Backend` interface and the file backend (`!js`): atomic put via temp file, fsync and rename; length- and CRC-framed append; and list/delete. Verify the unit tests cover atomic replace and dropping a torn trailing record.
- [x] 3.2 Implement the IndexedDB backend (`js`) with the same contract, including the `navigator.storage.persist()` request. Verify `GOOS=js GOARCH=wasm go vet ./...` passes. A shared backend contract test suite runs against the file backend in `go test`, and against IndexedDB later in the web integration check (task 12.2).
- [x] 3.3 Implement generations: snapshot plus journal replay, compaction thresholds (2,000 entries / 4 MB), monthly revlog segments, media blobs, the `current` pointer flip, orphan media GC, and reporting a `corrupt` status without deleting data. Verify the tests cover a kill-during-save simulation, a failed new generation leaving the old one intact, and opening a 50k-card / 500k-revlog synthetic collection in under 3 s.

## 4. Card rendering engine

- [x] 4.1 Implement the template parser and renderer in `render/`: fields, unknown-field markers, `FrontSide`, `#`/`^` sections, and the `text:`, `hint:` and `type:` filters. Verify golden tests cover every card-rendering spec scenario for these features.
- [x] 4.2 Implement cloze rendering (question/answer sides, hints, other clozes revealed), plus cloze-number extraction for card generation. Verify golden tests cover the cloze scenarios.
- [x] 4.3 Implement HTML sanitizing (strip scripts, event attributes, `[sound:]`, audio and video) and CSS inlining of `.card`, element, class and `.cloze` rules, with explicit background and colour extraction. Verify the tests cover the script, sound-tag and dark-mode-default scenarios.

## 5. Collection core

- [x] 5.1 Implement the domain model and the monotonic ID generator: decks (with `::` hierarchy and case-insensitive sibling uniqueness), deck options and their validation, built-in note types, and app settings. Implement every mutation as a journaled op. Verify the tests cover the deck spec scenarios (nested create, duplicate rejected, Default deck, last deck undeletable, retention out of range).
- [x] 5.2 Implement notes and card generation for standard and cloze types (using `render/`), regeneration on edit that preserves surviving cards, empty-note rejection, tags normalisation, and note type edits (rename and add/remove fields and templates). Verify the tests cover the reversed, cloze, edit-removes-cloze, empty-front, tag-splitting and remove-field scenarios.
- [x] 5.3 Implement deck deletion (cascade, orphan-note cleanup, card count for confirmation), moving cards, suspend/unsuspend, bury-until-rollover, and the change-event broadcaster. Verify the tests cover each collection scenario, and that change events fire for every op.

## 6. Scheduling

- [x] 6.1 Wrap go-fsrs in `scheduler/`: card states, memory-state updates, retention-targeted intervals, lapses, learning and relearning steps, Easy graduation, and deterministic fuzz. Verify the tests cover the scheduling spec scenarios (retention ordering, rating ordering, lapse, Good twice graduates, preview equals result).
- [x] 6.2 Implement the review log entries with pre-answer state, plus undo. Implement the day-boundary computation with a configurable rollover hour. Verify the tests cover the 02:30 / 04:00 scenario and exact-restore undo.
- [x] 6.3 Implement due counts and the queue builder: learning before limits, review and new limits including the parent/child rule, new cards in insertion order, sibling burying, and interleaving of new and review cards. Verify the tests cover the new-limit-reached, learning-ignores-limits and sibling-buried scenarios.

## 7. Search

- [x] 7.1 Implement the query parser and evaluator in `search/`: free text, `deck:`, `tag:`, `is:*`, `note:`, negation and quoted phrases, with precomputed stripped text per note and sorting by created, due and question. Verify the tests cover the combined-search and invalid-qualifier scenarios, and a 50k-card benchmark under 50 ms.

## 8. Import and backup engines

- [x] 8.1 Implement CSV/TSV staging and preview: delimiter and BOM detection, quoted multiline fields, header row, and column-to-field/tag mapping. Implement duplicate handling (skip/update/duplicate) as a single all-or-nothing journal batch. Verify the tests cover the tab-separated, quoted-fields and skip-duplicates scenarios.
- [x] 8.2 Implement the `.apkg` importer. It reads decks, note types (legacy JSON and `anki21b` protobuf), notes with GUIDs, cards, tags and image media (legacy JSON map and zstd protobuf `MediaEntries`). Audio is skipped with warnings, scheduling is converted according to the keep/reset choice, GUID re-import updates notes, and the result is committed atomically. Verify fixture tests cover modern, legacy, keep, reset, import-twice, summary-warnings and corrupt-package scenarios.
- [x] 8.3 Implement `.cardflip` export (manifest, collection, revlog, media) and restore (validate the manifest and format version, build a new generation, flip the pointer). Verify the tests cover the round-trip equality, wrong-file and newer-format scenarios.
- [x] 8.4 Add the `internal/yield` helpers (`_js` sleep, native no-op) and call them in import, export and compaction loops. Verify `GOOS=js GOARCH=wasm go vet ./...` passes, and a test checks that yield is called at the expected cadence.

## 9. Bridge services

- [x] 9.1 Implement the façades and DTOs in `golang/svc_*.go` for `app`, `decks`, `notes`, `study`, `browse`, `transfer` and `media` (design D2/D3/D7), with `Open(dataDir)` gating and `Changes()` and import-progress streams. Bind them in `fluffy.Start`. Verify `fluffy generate` emits no warnings, `go test ./...` covers each façade against an in-memory backend, and `tool/check.sh` passes.

## 10. Flutter app shell

- [x] 10.1 Implement `main.dart` bootstrap: resolve the data dir, call `golang.app.open`, and show a splash, then an error screen (offering restore or a new collection after confirmation), or the app. Implement repositories and `ChangeNotifier`s fed by `golang.app.changes()`. Verify widget tests with `Bridge.fake()` cover the ok, fresh and corrupt startup paths.
- [x] 10.2 Implement the theme (light/dark/system, persisted through Go settings), the adaptive scaffold (rail at ≥ 840 px, bottom bar below) with Decks, Browse, Add and Settings, and the global error snackbar handling. Verify widget tests cover the phone and wide layouts, and that the theme choice is sent to Go.

## 11. Flutter features

- [x] 11.1 Build the deck list home: tree with collapsible parents, coloured new/learning/review counts, deck actions (study, add, options, rename, delete with card-count confirmation), the empty-state first-run hint, and a deck options dialog. Verify widget tests cover the due-counts, empty-state and delete-confirmation scenarios.
- [x] 11.2 Build `CardView`: HTML rendering with inlined styles, a theme fallback for colours, and an image factory via `golang.media.get` with an LRU cache and a placeholder. Verify widget tests render a fake card with an image, a missing image and dark mode.
- [x] 11.3 Build the study screen: counts, reveal, rating buttons with previews, timing, undo, edit/suspend/bury actions, the nothing-due screen with next due time, the completion summary, and the keyboard shortcuts (Space/Enter, 1–4, Ctrl/Cmd+Z, Esc). Verify widget tests cover the reveal, rate, undo, keyboard and completion scenarios.
- [x] 11.4 Build the note editor: note type and deck pickers remembered between uses, field inputs, a tags input, rapid entry after a successful add, an error that keeps the user's input, and a note type manager (rename, fields, templates). Verify widget tests cover the rapid-entry, empty-note-error and save-failure scenarios.
- [x] 11.5 Build the browser: search field with error display, a paged lazy list (100-row pages), sorting, a details pane (preview, scheduling info, editor), and multi-select bulk actions with confirmation. Verify widget tests cover the search, edit and bulk-move flows. Manually confirm smooth scrolling with the synthetic 50k collection on Linux.
- [x] 11.6 Build the import flow: file picker (`stagePath` on native, `stageBytes` on web), CSV preview and mapping screen, `.apkg` options (keep/reset scheduling), a progress view from the import stream, and a summary or error. Verify widget tests cover the CSV mapping, `.apkg` options, progress and summary views.
- [x] 11.7 Build Settings: theme, rollover hour, import, export backup (`saveFile` / web download named `cardflip-<YYYY-MM-DD>.cardflip`), restore with an overwrite confirmation, the one-time persistent-storage notice on web, and About with the version. Verify widget tests cover the restore confirmation and the rollover change.
- [x] 11.8 Write `README.md` covering prerequisites (Go, FVM 3.47.0, fluffy on `PATH`, platform toolchains), how to run on each platform with `fluffy run`, `tool/check.sh`, the fixture regeneration script, and the supported Anki template/CSS subset. Verify the documented commands run as written on Linux.

## 12. Integration checks

- [x] 12.1 Add a Linux `integration_test` against the real Go library. It runs first launch, adds Basic and Cloze notes, studies with undo, imports the modern `.apkg` fixture with media, exports a backup, restores it, and restarts to confirm persistence. Verify it passes with `fvm flutter test integration_test -d linux`.
- [x] 12.2 Run the same flow on the web (`fluffy lib web`, then `flutter drive … -d web-server`), including the IndexedDB backend contract and a reload to confirm persistence. Verify it passes in Chrome.
- [x] 12.3 Build every target that is available on this machine (`fluffy build linux`, `fluffy build web`, `fluffy build apk` if the Android NDK is installed) and smoke-test each one manually. Record which targets were not built here, such as Windows, macOS and iOS, and what is needed to build them. Verify each build completes and the app opens to the deck list.
