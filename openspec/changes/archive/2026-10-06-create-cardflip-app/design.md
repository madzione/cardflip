# Design

## Context

This is a greenfield project. The repository holds only the OpenSpec setup and the fluffy skill. For motivation and scope, see proposal.md; for behaviour, see the specs. The environment imposes these constraints:

- **Toolchain**: Go 1.27, fluffy 0.1.0 in `~/go/bin` (which is not on the default `PATH` in this shell, but must be for builds), and Flutter 3.47.0 through FVM only. There is no global `flutter`, so the project is pinned with `.fvmrc`, and fluffy runs `fvm flutter` itself.
- **fluffy rules that shape the architecture**:
  - Native platforms load Go as a library and never run `main`.
  - The generator type-checks as `GOOS=js GOARCH=wasm`, so every exported API must build for wasm.
  - Go wasm runs on the browser's main thread and only yields when goroutines block.
  - Large payloads cross the bridge as JSON, with `[]byte` encoded as base64.
- **Platforms**: All six targets are required (see the app-shell spec). Every Go dependency must therefore be pure Go and compile for `js/wasm`, so cgo SQLite cannot be used.

## Goals / Non-Goals

**Goals:**
- All domain logic, including model, scheduling, rendering, search, import, backup and persistence, lives in plain Go packages. None of them import fluffy, so `go test` alone covers them.
- Dart is a thin UI layer over generated bindings, and its widget tests run against `Bridge.fake()`.
- A single persistence design works on native platforms (files) and the web (IndexedDB), behind one Go interface.

**Non-Goals:**
- A general SQL database. The collection is held in memory, because 50k cards fit easily.
- Full Anki template or CSS fidelity. Only the subset in the card-rendering spec is supported.
- Running the FSRS optimizer. Default parameters are used.

## Decisions

### D1. Bootstrap with `fluffy create` at the repo root, pinned via FVM

Run `fluffy create --fvm 3.47.0 --patch-gtk --org dev.cardflip --project-name cardflip --platforms linux,windows,macos,android,ios,web .` from the repo root, so `openspec/` and `.claude/` stay alongside the app. If fluffy refuses a non-empty directory, fall back to `fvm use 3.47.0`, then `fvm flutter create … .`, then `fluffy init --patch-gtk`.
- `--patch-gtk` is cheap and avoids the thick GTK title bar on non-GNOME desktops.
- *Alternative considered*: an `app/` subfolder. Rejected, because it complicates fluffy's project discovery and CI paths for no gain.

### D2. Go owns all state; services bound per area

`golang/main.go` sets up an unopened `core.Collection` inside `fluffy.Start` and binds thin service façades:

| Namespace | Responsibility |
|---|---|
| `app` | `Open(dataDir)`, `Status()`, settings (theme, rollover hour), `Changes() <-chan Change` |
| `decks` | deck tree with counts, CRUD, options |
| `notes` | note types, add/update/delete notes, tags |
| `study` | `Next(deckId)`, `Answer(cardId, rating, durationMs)`, `Undo()`, `Bury`, `Suspend`, previews |
| `browse` | `Search(query, sort) []int64`, `Rows(ids)`, `Card(id)`, bulk actions |
| `transfer` | stage/inspect/import files, export/restore backup |
| `media` | `Get(name) []byte` for card images |

The façades (in `golang/`) translate domain types into bridge DTOs, so domain packages never depend on JSON shapes. One `sync.Mutex` serializes all mutations; reads take an `RWMutex` read lock.
- *Alternative considered*: one big `collection` namespace. Rejected, because the generated Dart class would become unwieldy.

### D3. Explicit `Open(dataDir)` instead of opening in `fluffy.Start`

Go cannot discover the app data directory on Android or iOS, and `Start` runs before Dart can provide it. Dart therefore resolves the directory with `path_provider` (`getApplicationSupportDirectory()`, or `""` on web) and calls `golang.app.open(dir)` before showing any screen.
- Every other method returns a `not opened` error until that call succeeds.
- `Open` is idempotent across hot restarts.
- `Open` returns a status of `ok`, `fresh` or `corrupt`, which drives the error screen required by the persistence spec.

### D4. Persistence: in-memory model + snapshot + append-only journal, behind a `storage.Backend`

The collection, which holds everything except the review log and media, is loaded fully into memory. On disk it lives in a **generation**:
- `snapshot`: the full collection as JSON.
- `journal`: one JSON operation per entry, appended and fsynced per mutation.

On open, the app loads the snapshot and replays the journal. Compaction writes a new snapshot and clears the journal at open, and whenever the journal exceeds 2,000 entries or 4 MB.

- **Review log**: Stored separately, in monthly append-only segments (`revlog/2026-10`). Only the current month is loaded at open, because that is all that daily counts need. This keeps startup under 3 s with 500k entries.
- **Media**: Stored as separate blobs keyed by file name, and written before the operation that references them. Orphans are garbage-collected at compaction.
- **Atomic replace for import and restore**: The app builds a new generation (`gen-<n+1>`), writes it fully, and then flips a small `current` pointer record. A failure before the flip leaves the old generation untouched, which satisfies "failed import/restore changes nothing". Imports that only add data write one journal entry containing the whole batch, which is applied all-or-nothing on replay. Restores always use a new generation.
- **`storage.Backend`** is a small key/value interface with `Get`, `Put`, `Append`, `List(prefix)`, `Delete`, plus atomic `Put` semantics:
  - `backend_files.go` (`//go:build !js`): files under `dataDir`, written to a temp file, fsynced, then renamed. `Append` uses `O_APPEND` plus fsync. A torn last journal line is detected by a length and CRC prefix and dropped.
  - `backend_idb_js.go` (`//go:build js`): IndexedDB through `syscall/js`. Each journal entry is its own key (`journal/<seq>`). Promises are awaited by blocking on a Go channel, which is safe because fluffy runs handlers in goroutines. `navigator.storage.persist()` is requested at open, and its result is reported in `Status()`.
- *Alternatives considered*:
  - A pure-Go SQLite such as `modernc.org/sqlite`. Rejected, because it does not support `js/wasm`.
  - Rewriting one JSON file per change. Rejected, because it is too slow at 50k cards and stresses mobile flash.
  - Persisting from Dart through a `//fluffy:dart` interface. Rejected, because it would double the storage code across Dart and Go.

### D5. FSRS via `go-fsrs`, with our own learning-step layer and deterministic fuzz

`github.com/open-spaced-repetition/go-fsrs` (pure Go) provides the memory-state math, with default parameters and the deck's `request_retention`. Learning and relearning steps (1m/10m and 10m) are handled by `scheduler/` itself, the way Anki does it, so that the steps stay configurable regardless of the library's short-term mode. Interval fuzz is seeded from `cardId + reps`, so the preview equals the actual result, as the scheduling spec requires.
- Undo stores the pre-answer card state in the review log entry. `Undo` restores that state and deletes the entry, and both are journaled.
- `Due` queries take `now` and the rollover hour. The local zone comes from `time.Local`; on `js/wasm`, Go derives it from the browser's offset.

### D6. Template rendering in Go; HTML shown with `flutter_widget_from_html_core`

`render/` implements the card-rendering subset: fields, `FrontSide`, `#`/`^` sections, and the filters `text:`, `hint:`, `type:` and `cloze:`. Its output is sanitized HTML, from which `<script>`, event attributes, `[sound:]`, `<audio>` and `<video>` are removed.
- The note-type CSS is parsed with `github.com/tdewolff/parse/v2/css` (pure Go). Simple selectors (`.card`, element, class and `.cloze`) are inlined as `style=` attributes, because the Flutter HTML widget does not apply `<style>` blocks.
- `.card` background and colour are returned separately, so that Dart can fall back to theme colours when they are unset (theme-aware display).
- Dart resolves `<img src>` through a custom image factory that calls `golang.media.get(name)`, with an in-memory LRU cache and a placeholder on error.
- `flutter_widget_from_html_core` was chosen over a WebView because it is pure Dart, works on all six platforms, never executes scripts, and avoids platform WebView setup.

### D7. Import: staged sources, streaming progress, pure-Go `.apkg` reader

- **Staging**: Dart stages a file once and gets back a stage id. On native platforms it calls `transfer.stagePath(path)`, so Go reads the file directly and no bytes cross the bridge. On the web it calls `transfer.stageBytes(name, data)`. Then `inspect(id)` returns either a CSV preview or `.apkg` metadata, and `import(id, options)` returns `<-chan ImportEvent`, which arrives in Dart as a `Stream` of progress events followed by a summary or an error.
- **Web responsiveness**: Hot loops call `yield()`, which is `time.Sleep(time.Millisecond)` every N items in a `_js.go` file and a no-op on native.
- **CSV**: Handled with `encoding/csv` and `LazyQuotes`. The delimiter is detected by counting delimiter consistency over the first 20 lines, and a BOM is stripped.
- **`.apkg` steps**:
  1. Open the zip with `archive/zip`.
  2. Pick the collection file, preferring `collection.anki21b` (decompressed with zstd from `github.com/klauspost/compress`), then `collection.anki21`, then `collection.anki2`.
  3. Read the SQLite database with a minimal in-repo, read-only B-tree reader in `importer/apkg/sqlite` that works on an in-memory `[]byte`. It only does full table scans, covering rowid tables and `WITHOUT ROWID` tables (index B-trees), with overflow pages. Only `col`, `notes`, `cards`, `notetypes`/`fields`/`templates` and `decks` are read.
     - *Decision (task 1.3)*: `github.com/alicebob/sqlittle` was planned, but it fails the `js/wasm` requirement. Its only pager opens files through mmap and `golang.org/x/sys/unix` flock, which do not build for `js`, and it cannot read from memory, which is what the web's `stageBytes` path has. The fallback named in the risks below was therefore taken up front.
  4. Read note types: legacy packages keep them as JSON in `col.models`/`col.decks`. In `anki21b`, they are protobuf `config` blobs, decoded with `google.golang.org/protobuf/encoding/protowire` against the field numbers in Anki's `notetypes.proto` and `decks.proto`.
  5. Read media: the legacy `media` file is a JSON map. In `anki21b`, it is a zstd-compressed protobuf `MediaEntries`, and each media file is zstd-compressed individually.
- **Scheduling conversion** when the user keeps scheduling:
  - Review due date = `col.crt` + `due` days, converted to an absolute time.
  - Learning due is epoch seconds.
  - Memory state comes from `cards.data` `{"s","d"}` when present. Otherwise stability ≈ interval in days, and difficulty is mapped linearly from ease 130%–350% to 10–1, clamped.
- **Identity**: Notes are matched by Anki GUID, stored on the note, for re-import updates. Anki IDs are not reused, because new IDs come from our generator (D9).

### D8. Backup format `.cardflip`

The file is a zip containing:
- `manifest.json`: `{format: 1, appVersion, createdAt}`
- `collection.json`: the snapshot
- `revlog.jsonl`
- `media/<name>`

Export streams the zip into a byte buffer. Dart saves it with `file_picker`'s `saveFile`, which downloads the file on the web. Restore stages the file like an import, validates the manifest and every JSON part, builds a new generation and flips `current` (D4).

### D9. IDs and time on the wire

IDs are `int64` values derived from epoch milliseconds, bumped to stay strictly monotonic. They remain below 2^53, so they are exact on the web as well. Times cross the bridge as `time.Time`, which arrives in Dart as `DateTime` in UTC, and Dart formats them in local time.

### D10. Search in memory

`search/` parses the query into an AST, covering terms, qualifiers, negation and quoted phrases, and evaluates it by a linear scan over cards. The scan uses a per-note, precomputed, lowercased and HTML-stripped field text that is updated on edit. `Search` returns only ordered IDs, and Dart's `ListView.builder` fetches `Rows(ids[a:b])` in pages of 100. A linear scan of 50k cards takes milliseconds, so no index is built.

### D11. Dart structure, without a state-management framework

```
lib/
  main.dart            bootstrap: open collection → CardflipApp
  app/                 theme, adaptive scaffold (NavigationRail ≥ 840 px, NavigationBar below), routes
  data/                thin repositories wrapping golang.* + ChangeNotifiers fed by golang.app.changes()
  features/decks|study|browse|editor|import|settings/
  widgets/card_view.dart   HTML card renderer + media image factory
```

- `ChangeNotifier` with `ListenableBuilder` is enough for this size of app and adds no dependency.
- Navigation uses the plain `Navigator` inside each destination.
- Keyboard shortcuts use `Shortcuts`/`Actions` on the study screen.
- Settings live in the Go collection, so they are included in backups.

### D12. Go module layout

```
golang/
  main.go            fluffy.Start (construct core, Bind services); main() { fluffy.Serve() }
  svc_*.go           façades per namespace + DTOs
  core/              collection model, ops, journal apply, queues, change events
  scheduler/         FSRS wrapper, steps, fuzz, previews, Anki-state conversion
  render/            templates, cloze, sanitize, CSS inline
  search/            query parser + evaluator
  storage/           Backend interface, files (!js), IndexedDB (js), generations
  importer/csv, importer/apkg (+ sqlite reader shim, protobuf decode), backup/
  internal/yield/    yield_js.go / yield_other.go
```

### D13. Verification gates

Every task ends green on:
- in `golang/`: `go vet ./...`, `GOOS=js GOARCH=wasm go vet ./...` and `go test ./...`
- `fluffy generate --check`
- `fvm flutter analyze` and `fvm flutter test`

These are collected in `tool/check.sh`. `.apkg` fixtures are produced once with the official `anki` Python package by a committed script (`tool/make_apkg_fixtures.py`), and the resulting files are committed under `golang/importer/apkg/testdata/`.

## Risks / Trade-offs

- [The in-repo SQLite reader may not handle every page layout in real Anki DBs] → It replaced `sqlittle`, which does not build for `js/wasm` (see D7). The task 2 spike checks it against the fixture packages. Its own tests cover multi-level B-trees, overflow pages, `WITHOUT ROWID` tables and added columns, and they check that corrupt or truncated files fail with an error instead of a panic.
- [`anki21b` protobuf field numbers may drift between Anki versions] → Decode defensively by skipping unknown fields, pin the fixtures to a known Anki version, and fail with a clear "unsupported Anki package version" error rather than importing garbage.
- [Large `.apkg` files on the web go through base64 JSON, roughly 2.7× memory] → Accept this for v1 and document a soft limit of about 200 MB. Native platforms use `stagePath` and have no such cost.
- [Go wasm blocks the UI during heavy imports or compaction] → Use `yield()` in hot loops and stream progress. Compaction runs at open, behind a splash, and after imports, behind the progress UI.
- [IndexedDB eviction in browsers] → Request `persist()` and show the one-time backup advice from the persistence spec.
- [The in-memory model limits collection size] → 50k cards is about 50–100 MB of heap, which is acceptable. Much larger collections are a later concern.
- [The Flutter HTML widget does not support all of Anki's CSS] → Inline the common selectors and document the rest as unsupported. Cards stay readable even when styling is partial.
- [Learning steps diverging from go-fsrs's own short-term mode] → Unit-test step transitions against the scheduling-spec scenarios, and keep the memory-state update in the library.

## Migration Plan

This is a new app, so no data needs migrating. The storage format carries a version in the `current` pointer record, and backups carry `format: 1`, so later changes can migrate forward.

## Open Questions

- The final bundle identifier and organization (`dev.cardflip` is assumed) and the app icon can be changed before the first store release without affecting the specs.
- The default learning steps (1m/10m) could be exposed in deck options later. That would be an additive change.
