# Tasks

## 1. Settings

- [x] 1.1 Add `UIScale`, `Minimalist` and `ReduceMotion` to `core.Settings` with validated setters and DTO fields. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. App

- [x] 2.1 Apply the UI size in `lib/app/app.dart` through `Appearance` (`lib/app/appearance.dart`). It overrides `MediaQuery` size, pixel ratio and insets for a screen that many times smaller and draws the app with `Transform.scale`. Add a `test/app_shell_test.dart` test that 150 % enlarges a known text and icon. Verify it passes.
- [x] 2.2 Add minimalist variants to `lib/app/theme.dart`. Add a test that elevation is 0 on cards in minimalist mode. Verify it passes.
- [x] 2.3 Add reduce-motion helpers (from `MediaQuery.disableAnimations` or the setting) and use them for page transitions, the question fade, scroll animations, the theme animation and every dialog and menu (`popupAnimation`). Add tests that transitions and dialogs are instant when on, and that "follow system" follows the OS. Add a source check that every `showDialog`, `showMenu` and `PopupMenuButton` passes `popupAnimation`. Verify they pass.
- [x] 2.4 Add the three settings to `settings_screen.dart` with widget tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check 75 % and 200 % on a phone layout for overflow, covering the main destinations, studying and browser selection mode. Fix what overflows: settings rows with dropdowns, app bar actions, deck row counts, the browser mode toggle, the theme buttons, menu shortcut hints, and the context menu positions under scaling.
