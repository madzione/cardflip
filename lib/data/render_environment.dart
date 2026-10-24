import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'scope.dart';

/// Where cards are shown, which Go renders them for: the dark theme, the
/// platform classes and the card's display size. Go keeps one environment;
/// it is sent whenever it changes, and Go then tells open cards to render
/// again.
class RenderEnvironment {
  RenderEnvironment({required this.isWeb});

  final bool isWeb;

  /// Sizes are rounded down to steps of this many pixels, so resizing a
  /// window renders the card again only now and then.
  static const sizeStep = 50;

  bool _night = false;
  String _platform = '';
  int _width = 0;
  int _height = 0;
  golang.Env? _sent;

  /// The theme brightness and platform, from [RenderEnvironmentReporter].
  void setTheme({required bool night, required String platform}) {
    _night = night;
    _platform = platform;
    _send();
  }

  /// The size of the card view, from [CardView].
  void setCardSize(Size size) {
    _width = _bucket(size.width);
    _height = _bucket(size.height);
    _send();
  }

  static int _bucket(double px) =>
      px.isFinite && px > 0 ? px ~/ sizeStep * sizeStep : 0;

  void _send() {
    final sent = _sent;
    if (sent != null &&
        sent.night == _night &&
        sent.platform == _platform &&
        sent.widthPx == _width &&
        sent.heightPx == _height) {
      return;
    }
    final env = golang.Env(
      night: _night,
      platform: _platform,
      widthPx: _width,
      heightPx: _height,
    );
    _sent = env;
    unawaited(golang.app.setRenderEnvironment(env).catchError((Object _) {}));
  }
}

/// The platform classes Anki puts on cards: `linux`, `win`, `mac`,
/// `android mobile`, `iphone mobile` or `ipad mobile`. On the web only
/// `mobile`, on phones and tablets, so a deck looks the same in a desktop
/// browser on every operating system.
String platformClasses(
  TargetPlatform platform, {
  required bool isWeb,
  required Size screen,
}) {
  final mobile =
      platform == TargetPlatform.android || platform == TargetPlatform.iOS;
  if (isWeb) return mobile ? 'mobile' : '';
  return switch (platform) {
    TargetPlatform.linux => 'linux',
    TargetPlatform.windows => 'win',
    TargetPlatform.macOS => 'mac',
    TargetPlatform.android => 'android mobile',
    TargetPlatform.iOS =>
      screen.shortestSide >= 600 ? 'ipad mobile' : 'iphone mobile',
    TargetPlatform.fuchsia => '',
  };
}

/// Reports the effective theme brightness and the platform to the
/// [RenderEnvironment]; it sits below the app's theme.
class RenderEnvironmentReporter extends StatefulWidget {
  const RenderEnvironmentReporter({super.key, required this.child});

  final Widget child;

  @override
  State<RenderEnvironmentReporter> createState() =>
      _RenderEnvironmentReporterState();
}

class _RenderEnvironmentReporterState extends State<RenderEnvironmentReporter> {
  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final app = AppScope.read(context);
    app.renderEnv.setTheme(
      night: Theme.of(context).brightness == Brightness.dark,
      platform: platformClasses(
        defaultTargetPlatform,
        isWeb: app.isWeb,
        screen: MediaQuery.sizeOf(context),
      ),
    );
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
