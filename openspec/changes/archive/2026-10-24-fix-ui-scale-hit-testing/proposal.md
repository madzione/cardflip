# Proposal

## Why

At any UI size below 100 % (75–95 %), controls near the right and bottom edges of the window still show up but do not respond to taps or clicks. At 75 % the dead strip covers the right and bottom quarter of the window. Users who pick a smaller size can get stuck: the UI size dropdown sits on the right side of Settings, so they cannot switch back to 100 %.

## What Changes

- Fix the UI size scaling so that pointer input (taps, clicks, hovers, drags, scrolls) reaches every visible element at every UI size from 75 % to 200 %, across the whole window.
- Make the UI size requirement say plainly that the app stays fully usable at every size, so this kind of regression shows up in tests.
- Add tests that tap controls at the window's right and bottom edges at sizes below 100 %, including switching the UI size from 75 % back to 100 %.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `app-shell`: The "UI size" requirement gains the guarantee that every visible element can still be used at any UI size, with a scenario for switching back from 75 % to 100 %.

## Impact

- `lib/app/appearance.dart`: the `Appearance` widget that scales the app (this is where the fix goes).
- `test/app_shell_test.dart`, `test/settings_test.dart`: new regression tests.
- No change to persistence, the Go API (`app.SetUIScale`) or the settings range. Users stuck at a small size are unblocked once they update.
