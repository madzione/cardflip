# Tasks

## 1. Regression tests (fail before the fix)

- [x] 1.1 In `test/app_shell_test.dart`'s `UI size` group, set `WidgetController.hitTestWarningShouldBeFatal = true` (reset in `tearDown`). Run `flutter test test/app_shell_test.dart` and confirm that "75 % fits a phone on every destination" now fails, because taps on the bottom `NavigationBar` miss
- [x] 1.2 Add an `Appearance` widget test that wraps a full-window `Listener` (or `GestureDetector`) and, at 75 %, 90 %, 100 % and 150 %, taps the four corners and the center of the window. Check that each tap arrives at the expected unscaled local position (`screenPoint / scale`). Confirm it fails at 75 % and 90 % before the fix
- [x] 1.3 In `test/settings_test.dart`'s `appearance` group, add a test for the spec scenario "Switching back from a smaller size": start with `uiScale: 75` on a desktop-sized window, open Settings, open the `ui-size` dropdown and pick `100%`, with fatal hit-test warnings on. Check that `app.SetUIScale` was called with `[100]` and that the dropdown shows 100. Confirm it fails before the fix

## 2. Fix

- [x] 2.1 In `lib/app/appearance.dart`, replace `Transform.scale` + `OverflowBox` with `FittedBox(fit: BoxFit.fill, alignment: Alignment.topLeft, child: SizedBox.fromSize(size: size, child: child))`. Keep the same widget structure at every scale, and update the class doc comment to describe the approach and why taps reach the whole window. Verify that tests 1.1–1.3 now pass
- [x] 2.2 Verify the existing UI size tests still pass unchanged ("150 % makes text and icons 1.5 times larger at once (spec)", the 200 % phone tests, and the settings `appearance` tests) with `flutter test test/app_shell_test.dart test/settings_test.dart`

## 3. Integration checks

- [x] 3.1 Run the whole widget test suite (`flutter test`) and verify there are no new failures, in particular in tests that check widget rects or semantics under `Appearance`
- [x] 3.2 Manual check on Linux desktop (`flutter run -d linux`): set the UI size to 75 %, then click controls at the right edge (the Settings dropdowns, app bar actions) and bottom edge, scroll a list on the right half, and switch back to 100 %. Repeat at 90 % and 150 %. Verify everything responds, and that menus and tooltips open at the pointer
