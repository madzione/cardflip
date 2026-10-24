import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';

import 'package:cardflip/data/profiles_model.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/sync_model.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/features/profiles/profile_list_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/builtin_names.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/widgets/font_cache.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/media_files.dart';

import 'app.dart';
import 'appearance.dart';
import 'errors.dart';
import 'theme.dart';

/// Where the collection is stored: the app support directory on native
/// platforms; on the web Go uses IndexedDB and needs no directory.
Future<String> defaultDataDir() async =>
    kIsWeb ? '' : (await getApplicationSupportDirectory()).path;

/// Opens the collection of the last used profile, then shows the app — or
/// the error screen when the stored data cannot be read, or first the
/// profile list when "Show profile list at startup" is on.
class CardflipBootstrap extends StatefulWidget {
  const CardflipBootstrap({
    super.key,
    this.dataDir = defaultDataDir,
    this.isWeb = kIsWeb,
    this.picker = const FilePickerService(),
    this.syncPlatform,
  });

  final Future<String> Function() dataDir;
  final bool isWeb;
  final FilePickerService picker;

  /// The platform services device sync uses; tests fake them.
  final SyncPlatform? syncPlatform;

  @override
  State<CardflipBootstrap> createState() => _CardflipBootstrapState();
}

class _CardflipBootstrapState extends State<CardflipBootstrap> {
  golang.Status? _status;
  Object? _error;
  AppState? _app;

  /// The profiles, for the profile list shown before the app.
  final _profiles = ProfilesModel();

  /// Device sync, the same for every profile.
  late final _sync = SyncModel(
    isWeb: widget.isWeb,
    platform: widget.syncPlatform,
  );
  bool _chooseProfile = false;

  @override
  void initState() {
    super.initState();
    _open();
  }

  Future<void> _open() async {
    setState(() => _error = null);
    try {
      // A new collection has no language setting yet, so its deck and note
      // types are named in the device's language. The app sends the
      // setting's names once it runs (L10nTracker).
      await sendBuiltinNames(
        lookupAppLocalizations(
          resolveLocale(
            WidgetsBinding.instance.platformDispatcher.locales,
            AppLocalizations.supportedLocales,
          ),
        ),
      ).catchError((Object _) {});
      final dataDir = await widget.dataDir();
      _profiles.list = await golang.app.loadProfiles(dataDir);
      // After a hot restart a profile is open already.
      if (_profiles.list.showAtStartup && _profiles.list.current.isEmpty) {
        setState(() => _chooseProfile = true);
        return;
      }
      final status = await golang.app.open(dataDir);
      if (status.state == 'corrupt') {
        setState(() => _status = status);
        return;
      }
      await _start(status);
    } catch (e) {
      setState(() => _error = e);
    }
  }

  Future<void> _start(golang.Status status) async {
    final app = AppState(
      isWeb: widget.isWeb,
      picker: widget.picker,
      switchProfile: _openProfile,
      sync: _sync,
    );
    await app.load();
    if (!mounted) {
      app.dispose();
      return;
    }
    final old = _app;
    setState(() {
      _status = status;
      _app = app;
      _chooseProfile = false;
    });
    _disposeLater(old);
    _sync.start();
  }

  /// Closes the open collection, if any, and opens profile [id]'s: the app
  /// starts over with a new [AppState].
  Future<void> _openProfile(String id) async {
    final status = await golang.app.switchProfile(id);
    // Media files of the same name differ between profiles.
    MediaCache.instance.clear();
    FontCache.instance.clear();
    StoredMediaFiles.forget();
    if (status.state != 'corrupt') return _start(status);
    await _profiles.load();
    final old = _app;
    setState(() {
      _status = status;
      _app = null;
      _chooseProfile = false;
    });
    _disposeLater(old);
  }

  /// Disposes [app] once the screens using it are gone.
  void _disposeLater(AppState? app) {
    if (app == null) return;
    WidgetsBinding.instance.addPostFrameCallback((_) => app.dispose());
  }

  Future<void> _showProfiles() async {
    await _profiles.load();
    setState(() => _chooseProfile = true);
  }

  @override
  void dispose() {
    _app?.dispose();
    _profiles.dispose();
    _sync.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final app = _app;
    if (app != null) {
      return AppScope(
        state: app,
        // A new app for each opened collection: its screens start over.
        child: CardflipApp(key: ObjectKey(app), status: _status!),
      );
    }
    final Widget home;
    if (_chooseProfile) {
      home = ProfileListScreen(
        profiles: _profiles,
        onOpen: _openProfile,
        atStartup: true,
      );
    } else if (_error != null) {
      final error = _error!;
      home = ErrorScreen(
        message: (l) => l.couldNotStart(errorMessage(error)),
        onRetry: _open,
      );
    } else if (_status?.state == 'corrupt') {
      final message = _status!.message;
      home = ErrorScreen(
        message: (l) =>
            message == null ? l.collectionUnreadableShort : goText(message, l),
        onRestore: () async {
          final staged = await widget.picker.pick(const ['cardflip']);
          if (staged == null) return;
          if (await guard(() => golang.transfer.restore(staged.id))) {
            await _start(await golang.app.status());
          }
        },
        onStartFresh: () async {
          final status = await golang.app.startFresh();
          await _start(status);
        },
        onSwitchProfile: _profiles.several ? _showProfiles : null,
      );
    } else {
      home = const SplashScreen();
    }
    // Before the collection is open its language setting is unknown, so
    // these screens follow the device's languages.
    return MaterialApp(
      title: 'Cardflip',
      debugShowCheckedModeBanner: false,
      scaffoldMessengerKey: messengerKey,
      theme: lightTheme(),
      darkTheme: darkTheme(),
      supportedLocales: AppLocalizations.supportedLocales,
      localeListResolutionCallback: resolveLocale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      builder: (context, child) => L10nTracker(child: child!),
      home: home,
    );
  }
}

class SplashScreen extends StatelessWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.style_outlined, size: 64),
            SizedBox(height: 16),
            Text('Cardflip'),
            SizedBox(height: 24),
            CircularProgressIndicator(),
          ],
        ),
      ),
    );
  }
}

/// Shown when the collection cannot be read. The data stays untouched until
/// the user restores a backup or explicitly starts over.
class ErrorScreen extends StatefulWidget {
  const ErrorScreen({
    super.key,
    required this.message,
    this.onRestore,
    this.onStartFresh,
    this.onRetry,
    this.onSwitchProfile,
  });

  /// The message, in the interface language.
  final String Function(AppLocalizations l) message;
  final Future<void> Function()? onRestore;
  final Future<void> Function()? onStartFresh;
  final Future<void> Function()? onRetry;

  /// Shows the profile list, to open another profile instead.
  final Future<void> Function()? onSwitchProfile;

  @override
  State<ErrorScreen> createState() => _ErrorScreenState();
}

class _ErrorScreenState extends State<ErrorScreen> {
  bool _busy = false;

  Future<void> _run(Future<void> Function() action) async {
    setState(() => _busy = true);
    await guard(action);
    if (mounted) setState(() => _busy = false);
  }

  Future<void> _confirmStartFresh() async {
    final ok = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => AlertDialog(
        title: Text(context.l10n.startNewCollectionTitle),
        content: Text(context.l10n.startNewCollectionBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            ),
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.deleteAndStartOver),
          ),
        ],
      ),
    );
    if (ok == true) await _run(widget.onStartFresh!);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l = context.l10n;
    return Scaffold(
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 480),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Icon(
                  Icons.error_outline,
                  size: 56,
                  color: theme.colorScheme.error,
                ),
                const SizedBox(height: 16),
                Text(
                  l.collectionCouldNotOpen,
                  style: theme.textTheme.headlineSmall,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 12),
                Text(widget.message(l), textAlign: TextAlign.center),
                const SizedBox(height: 8),
                if (widget.onStartFresh != null)
                  Text(
                    l.nothingDeleted,
                    textAlign: TextAlign.center,
                    style: theme.textTheme.bodySmall,
                  ),
                const SizedBox(height: 24),
                if (_busy) const Center(child: CircularProgressIndicator()),
                if (!_busy && widget.onRestore != null)
                  FilledButton.icon(
                    onPressed: () => _run(widget.onRestore!),
                    icon: const Icon(Icons.restore),
                    label: Text(l.restoreFromBackup),
                  ),
                if (!_busy && widget.onStartFresh != null) ...[
                  const SizedBox(height: 8),
                  OutlinedButton(
                    onPressed: _confirmStartFresh,
                    child: Text(l.startNewCollection),
                  ),
                ],
                if (!_busy && widget.onSwitchProfile != null) ...[
                  const SizedBox(height: 8),
                  TextButton(
                    onPressed: () => _run(widget.onSwitchProfile!),
                    child: Text(l.switchProfile),
                  ),
                ],
                if (!_busy && widget.onRetry != null)
                  FilledButton(
                    onPressed: () => _run(widget.onRetry!),
                    child: Text(l.tryAgain),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
