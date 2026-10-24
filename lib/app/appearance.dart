import 'dart:ui' show DisplayFeature;

import 'package:flutter/material.dart';

/// Applies the UI size and reduce motion settings to the app below it; it
/// sits below the app's theme, above the navigator.
///
/// The UI size lays the app out for a screen [scale] times smaller than the
/// real one and draws it [scale] times larger, so that text, icons and
/// spacing all grow alike, on top of the system's text scale. A [FittedBox]
/// stretches the app, laid out at the smaller size, over the real screen; it
/// maps pointer events into the app before checking them against the app's
/// own size, so taps reach the whole window at any scale, also below 1 where
/// the app is laid out larger than the screen.
class Appearance extends StatelessWidget {
  const Appearance({
    super.key,
    required this.scale,
    required this.reduceMotion,
    required this.child,
  });

  /// The UI size as a factor, 1 for 100 %.
  final double scale;

  /// The reduce motion setting: "system", "on" or "off".
  final String reduceMotion;

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final data = MediaQuery.of(context);
    final reduce = reduceMotionIn(reduceMotion, system: data.disableAnimations);
    final theme = Theme.of(context);
    final size = data.size / scale;
    // The same widgets whatever the settings, so that changing them keeps
    // the app's state.
    return MediaQuery(
      data: data.copyWith(
        size: size,
        devicePixelRatio: data.devicePixelRatio * scale,
        padding: data.padding / scale,
        viewPadding: data.viewPadding / scale,
        viewInsets: data.viewInsets / scale,
        systemGestureInsets: data.systemGestureInsets / scale,
        displayFeatures: [
          for (final f in data.displayFeatures)
            DisplayFeature(
              bounds: Rect.fromPoints(
                f.bounds.topLeft / scale,
                f.bounds.bottomRight / scale,
              ),
              type: f.type,
              state: f.state,
            ),
        ],
        disableAnimations: reduce,
      ),
      child: Theme(
        data: reduce
            ? theme.copyWith(pageTransitionsTheme: instantPageTransitions)
            : theme,
        child: FittedBox(
          fit: BoxFit.fill,
          alignment: Alignment.topLeft,
          child: SizedBox.fromSize(size: size, child: child),
        ),
      ),
    );
  }
}

/// Whether the screen is too narrow, as on a phone with a large UI size,
/// for more than a couple of action icons in an app bar (the others go into
/// its menu) and for shortcut hints in menus.
bool narrowScreen(BuildContext context) =>
    MediaQuery.sizeOf(context).width < 360;

/// Whether the reduce motion [setting] ("system", "on" or "off") is in
/// effect, when the system asks for reduced motion ([system]) or not.
bool reduceMotionIn(String setting, {required bool system}) =>
    switch (setting) {
      'on' => true,
      'off' => false,
      _ => system,
    };

/// Whether transitions and animations are replaced by immediate changes
/// ("Reduce motion").
bool reduceMotion(BuildContext context) =>
    MediaQuery.disableAnimationsOf(context);

/// [duration], or none when reduce motion is in effect.
Duration motionDuration(BuildContext context, Duration duration) =>
    reduceMotion(context) ? Duration.zero : duration;

/// The animation of a dialog or menu opened from [context]: none when
/// reduce motion is in effect, else its own.
AnimationStyle? popupAnimation(BuildContext context) =>
    reduceMotion(context) ? AnimationStyle.noAnimation : null;

/// Page transitions that show the new page at once, for reduce motion.
final instantPageTransitions = PageTransitionsTheme(
  builders: {
    for (final platform in TargetPlatform.values)
      platform: const _InstantPageTransitionsBuilder(),
  },
);

class _InstantPageTransitionsBuilder extends PageTransitionsBuilder {
  const _InstantPageTransitionsBuilder();

  @override
  Duration get transitionDuration => Duration.zero;

  @override
  Duration get reverseTransitionDuration => Duration.zero;

  @override
  Widget buildTransitions<T>(
    PageRoute<T> route,
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
    Widget child,
  ) => child;
}
