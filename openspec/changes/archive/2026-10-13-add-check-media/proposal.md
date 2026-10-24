# Proposal

## Why

Media files take most of a collection's space, especially on the web, where browser storage is limited. Cardflip deletes unreferenced media silently when it compacts storage, and never reports files that notes reference but that are missing. Anki's Check Media lists both, lets the user delete unused files into a trash that can be restored, and tags notes with missing media so they can be fixed.

## What Changes

- **Check media** in Settings reports:
  - **unused files**: media that no note field and no note type template or CSS refers to. Files whose names start with `_` are never reported as unused, because templates and CSS use that convention for static files;
  - **missing files**: names that notes refer to but that are not in the media.
- Actions in the report: **Delete unused** (moves them to the media trash), **Tag missing** (adds the tag `missing-media` to the notes that refer to missing files), **Empty trash**, and **Restore deleted** (moves trashed files back).
- Compaction no longer deletes unreferenced media directly; it moves them to the trash, so nothing is lost silently. Files in the trash for more than 30 days are deleted at compaction.
- The report shows the space used by unused files and by the trash.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `persistence`: new requirements for checking media, the media trash, and keeping unreferenced media recoverable.
- `app-shell`: new requirement for the Check media entry.

## Impact

- Go: media references from fields (`<img src>`, `[sound:]`, `<audio>`, `<video>`, CSS `url()`), templates and CSS in `golang/core` (extending `referencedMedia`); a trash area in `storage` (files backend: a `media.trash` directory; IndexedDB: a separate store) with move, restore, list and purge; compaction moves orphans to the trash; `AppService.CheckMedia`, `DeleteUnusedMedia`, `TagMissingMedia`, `EmptyMediaTrash`, `RestoreMediaTrash`.
- Flutter: `settings_screen.dart` (entry and report).
- Tests: `golang/core` media tests, `golang/storage` contract tests for the trash, `golang/svc_test.go`, `test/settings_test.dart`.
