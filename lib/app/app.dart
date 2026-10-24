import 'package:flutter/material.dart';
import 'package:flutter_quill/flutter_quill.dart'
    show FlutterQuillLocalizations;

import 'package:cardflip/data/render_environment.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/settings/sync_dialogs.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'appearance.dart';
import 'errors.dart';
import 'scaffold.dart';
import 'shortcuts.dart';
import 'theme.dart';

/// The app once the collection is open.
class CardflipApp extends StatefulWidget {
  const CardflipApp({super.key, required this.status});

  final golang.Status status;

  @override
  State<CardflipApp> createState() => _CardflipAppState();
}

class _CardflipAppState extends State<CardflipApp> {
  final _navigatorKey = GlobalKey<NavigatorState>();

  @override
  void initState() {
    super.initState();
    // Go asks about pairing and first syncs; on the web there is no sync.
    if (!AppScope.read(context).isWeb) {
      golang.SyncUI.register(SyncDialogs(() => _navigatorKey.currentContext));
    }
    if (widget.status.showPersistenceNotice) {
      WidgetsBinding.instance.addPostFrameCallback(
        (_) => _showPersistenceNotice(),
      );
    }
  }

  /// Once, on the web, when the browser may clear the site's data.
  Future<void> _showPersistenceNotice() async {
    final context = _navigatorKey.currentContext;
    if (context == null) return;
    await showDialog<void>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.warning_amber_outlined),
        title: Text(context.l10n.browserMayClear),
        content: Text(context.l10n.browserMayClearBody),
        actions: [
          FilledButton(
            onPressed: () => Navigator.pop(context),
            child: Text(context.l10n.ok),
          ),
        ],
      ),
    );
    await guard(golang.app.persistenceNoticeShown);
  }

  @override
  Widget build(BuildContext context) {
    final settings = AppScope.of(context).settings;
    return ListenableBuilder(
      listenable: settings,
      builder: (context, _) => MaterialApp(
        title: 'Cardflip',
        debugShowCheckedModeBanner: false,
        navigatorKey: _navigatorKey,
        scaffoldMessengerKey: messengerKey,
        theme: lightTheme(minimalist: settings.minimalist),
        darkTheme: darkTheme(minimalist: settings.minimalist),
        themeMode: settings.themeMode,
        // With "follow system", Flutter already shortens animations when
        // the system asks for reduced motion.
        themeAnimationDuration: settings.reduceMotion == 'on'
            ? Duration.zero
            : kThemeAnimationDuration,
        locale: localeOf(settings.language),
        supportedLocales: AppLocalizations.supportedLocales,
        localeListResolutionCallback: resolveLocale,
        localizationsDelegates: const [
          ...AppLocalizations.localizationsDelegates,
          // The rich field editor's dialogs and menus.
          FlutterQuillLocalizations.delegate,
        ],
        home: const AdaptiveScaffold(),
        // App-wide shortcuts, above the navigator so that they work whatever
        // route or dialog has focus. Below the theme, cards learn whether
        // they are shown in dark mode, and below the UI size, the size of
        // the screen they are shown on.
        builder: (context, child) => Appearance(
          scale: settings.uiScale / 100,
          reduceMotion: settings.reduceMotion,
          child: L10nTracker(
            child: RenderEnvironmentReporter(
              child: CallbackShortcuts(
                bindings: {
                  backupShortcut: AppScope.read(context).backups.create,
                },
                child: child!,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
