# Proposal

## Why

Cardflip's interface size follows only the OS text scale, which many desktop users never change and which does not scale icons or spacing. Anki offers a UI size setting, a minimalist mode with fewer visual effects, and a reduce-motion option for users sensitive to animation.

## What Changes

- **UI size** setting: 75 % to 200 % in steps of 5 % (default 100 %), on top of the OS text scale. The app is laid out for a screen that many times smaller and drawn scaled up, so text, icons (including fixed-size ones) and spacing all grow alike. A large size on a narrow window can switch to the phone layout. The main screens adapt to narrow widths, so 200 % fits a phone without overflow.
- **Minimalist mode**: flat surfaces without shadows or elevation tints, thinner dividers and no decorative colour fills.
- **Reduce motion**: page transitions, dialogs and menus opening, the question fade, scroll animations and the theme switch become instant. It follows the OS "reduce motion" preference by default and can be turned on or off explicitly. When the OS asks for reduced motion, Flutter also shortens its own animations. "Off" cannot undo that shortening.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-shell`: new requirements for UI size, minimalist mode and reduce motion.

## Impact

- Go: `core.Settings` gains `UIScale`, `Minimalist` and `ReduceMotion` (system/on/off) with setters.
- Flutter:
  - `lib/app/appearance.dart` (new):
    - the `Appearance` widget: a `MediaQuery` override of size, pixel ratio, insets and `disableAnimations`, plus `Transform.scale`;
    - instant page transitions;
    - the `reduceMotion`, `motionDuration`, `popupAnimation` and `narrowScreen` helpers.
  - `lib/app/app.dart` applies it and the theme animation duration.
  - `lib/app/theme.dart` gets minimalist variants of the light and dark themes.
  - Every `showDialog`, `showMenu` and `PopupMenuButton` passes `popupAnimation(context)`.
  - `settings_screen.dart`.
  - Narrow-screen layouts:
    - settings rows put their dropdown under the label;
    - the deck list and browser app bars, including selection mode, move actions into the ⋮ menu;
    - deck rows show their counts under the name;
    - the browser's Cards/Notes toggle and the theme buttons scale down;
    - menus hide shortcut hints.
  - The browser's column and sidebar context menus convert the pointer position into overlay coordinates.
- Tests:
  - `golang/svc_test.go`;
  - `test/settings_test.dart`;
  - `test/app_shell_test.dart`, including 75 % and 200 % on a phone;
  - `test/reduce_motion_popups_test.dart` (source check that every dialog and menu follows reduce motion).
- Overlap: `add-review-display-options` has a separate question fade; reduce motion turns that fade off as well.
