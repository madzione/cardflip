# Design

## Context

`Appearance` (`lib/app/appearance.dart`) applies the UI size. It lays the app out for a screen `size = screenSize / scale`, using a `MediaQuery` with that size, and draws it `scale` times larger:

```
Transform.scale(scale, topLeft)
  └─ OverflowBox(min/max = size)   ← its own size is the parent's: the real screen
       └─ app (laid out at `size`)
```

Hit testing goes through `RenderTransform`, which maps the pointer into the unscaled coordinates (`p / scale`). Then `OverflowBox`'s render object runs the default `RenderBox.hitTest`, which first checks `size.contains(p)` against **its own** size, the real screen size, not its child's.

- `scale > 1`: the child is smaller than the box, so every point on the child passes the check. That is why 150 % and 200 % work.
- `scale < 1`: the child (`screen / scale`) is larger than the box. Points on screen past `scale × width` or `scale × height` map to child coordinates beyond the box and get rejected before they reach the child. At 75 % that is the right and bottom 25 % of the window, which matches the report. Painting has no such check, so those controls still show.

The existing "75 % fits a phone" test doesn't catch this because `tester.tap` only prints a warning when the tap misses its target (`WidgetController.hitTestWarningShouldBeFatal` is `false` by default).

## Goals / Non-Goals

**Goals:**
- Pointer events reach the whole scaled app at every UI size, with painting, layout and `MediaQuery` values as they are now.
- Keep the widget structure the same at every UI size, so changing it keeps the app's state (the constraint stated in `Appearance`).

**Non-Goals:**
- Changing the UI size range, step or default, or how it is persisted.
- Reworking the UI size to use `textScaler`, `visualDensity` or theme scaling instead of a transform.

## Decisions

### Replace `Transform.scale` + `OverflowBox` with `FittedBox(fit: BoxFit.fill)` around a `SizedBox` of the layout size

```
FittedBox(fit: BoxFit.fill, alignment: Alignment.topLeft)
  └─ SizedBox.fromSize(size: size)
       └─ app
```

`RenderFittedBox` sizes itself to the real screen, lays the child out at the `SizedBox` size, and computes the transform from the two sizes (`screen / size = scale`). Its `hitTestChildren` applies that transform and passes the mapped point to the child, which checks it against the child's own size (the full layout size). Its own bounds check runs on the unmapped point, against the screen-sized box, so it always passes for points on screen. It also implements `applyPaintTransform`, so `localToGlobal`/`globalToLocal` (used by menus, tooltips and drag feedback) stay correct.

Alternatives considered:
- **Custom `RenderProxyBox` / `OverflowBox` subclass that overrides `hitTest` to skip its own size check.** It works, but it is a hand-written render object doing what `FittedBox` already does and tests.
- **Wrapping in something larger than the screen so the bounds check passes.** Parent constraints are tight to the screen size, so no box below `Transform` can be larger than the screen. Not possible.
- **Scaling with `textScaler` + icon theme + density.** It does not scale spacing and custom sizes alike, and it would change behavior the spec relies on ("text, icons and spacing all grow alike"). Rejected.

### Make missed taps fatal in the UI size tests

Set `WidgetController.hitTestWarningShouldBeFatal = true` for the UI size test group, reset in `tearDown`, so the existing 75 %/200 % phone tests fail if a tap misses. Add targeted tests on top of that (see tasks).

## Risks / Trade-offs

- [`BoxFit.fill` scales width and height independently, so rounding in `size / scale` could give a slightly different factor per axis] → The difference is sub-pixel and has no visible effect. Tests compare sizes with a `closeTo` tolerance, as they already do.
- [A zero-sized window (minimised on desktop) gives an empty `FittedBox`] → `RenderFittedBox` already skips painting and hit testing when either size is empty, so nothing breaks. The old code did the same in effect.
- [Different render objects in the tree can change semantics or golden output] → The paint transform is the same (scale about the top left), so pixels are unchanged. Run the full widget test suite to catch semantics or rect differences.
