import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/licenses.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/settings_model.dart' show maxRestDaysPerWeek;
import 'package:cardflip/features/export/export_dialog.dart';
import 'package:cardflip/features/import/import_screen.dart';
import 'package:cardflip/features/profiles/profile_list_screen.dart';
import 'package:cardflip/features/settings/daily_goal_dialog.dart';
import 'package:cardflip/features/settings/sync_section.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/widgets/font_cache.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/media_files.dart';
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  /// Opens the repository in the system browser, or a new tab on the web.
  @visibleForTesting
  static Future<bool> Function(Uri url) openUrl = (url) =>
      launchUrl(url, mode: LaunchMode.externalApplication);

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  late final Future<golang.Status> _status = golang.app.status();
  bool _busy = false;

  Future<void> _run(Future<void> Function() action) async {
    setState(() => _busy = true);
    await guard(action);
    if (mounted) setState(() => _busy = false);
  }

  Future<void> _export() {
    final picker = AppScope.of(context).picker;
    return _run(() async {
      final file = await golang.transfer.exportBackup();
      if (await picker.save(file.name, file.data)) {
        showInfo(appL10n.backupSavedAs(file.name));
      }
    });
  }

  Future<void> _restore() async {
    final picker = AppScope.of(context).picker;
    golang.StageInfo? staged;
    golang.BackupInfo? backup;
    final read = await guard(() async {
      staged = await picker.pick(const ['cardflip']);
      if (staged == null) return;
      backup = (await golang.transfer.inspect(staged!.id)).backup;
      if (backup == null) {
        await golang.transfer.discard(staged!.id);
        throw golang.GolangException('transfer.Inspect', appL10n.goNotBackup);
      }
    });
    if (!read || staged == null || backup == null || !mounted) return;
    final created = backup!.createdAt.toLocal();
    final ok = await confirm(
      context,
      title: context.l10n.replaceCollectionTitle,
      message: context.l10n.restoreBackupBody(formatDate(created)),
      action: context.l10n.restore,
      destructive: true,
    );
    if (!ok) {
      await guard(() => golang.transfer.discard(staged!.id));
      return;
    }
    await _run(() async {
      await golang.transfer.restore(staged!.id);
      MediaCache.instance.clear();
      FontCache.instance.clear();
      StoredMediaFiles.forget();
      showInfo(appL10n.backupRestored);
    });
  }

  Future<void> _exportDeletionLog() {
    final picker = AppScope.of(context).picker;
    return _run(() async {
      final file = await golang.app.exportDeletionLog();
      if (file.notes == 0) {
        showInfo(appL10n.deletionLogEmpty);
      } else if (await picker.save(file.name, file.data)) {
        showInfo(appL10n.deletionLogSaved(file.notes, file.name));
      }
    });
  }

  Future<void> _clearDeletionLog() async {
    final ok = await confirm(
      context,
      title: context.l10n.clearDeletionLogTitle,
      message: context.l10n.clearDeletionLogBody,
      action: context.l10n.clear,
      destructive: true,
    );
    if (!ok) return;
    await _run(() async {
      await golang.app.clearDeletionLog();
      showInfo(appL10n.deletionLogCleared);
    });
  }

  Future<void> _renameFlag(int flag) async {
    final settings = AppScope.read(context).settings;
    final name = await askText(
      context,
      title: context.l10n.renameFlag(flagColorName(flag)),
      initial: settings.flagNames[flag - 1],
      label: context.l10n.name,
      helper: context.l10n.renameFlagHelper(flagColorName(flag)),
    );
    if (name != null) await settings.setFlagName(flag, name);
  }

  Future<void> _editDefaultSearch() async {
    final settings = AppScope.read(context).settings;
    await askAndSave(
      context,
      title: context.l10n.defaultBrowserSearch,
      initial: settings.defaultSearch,
      label: context.l10n.search,
      hint: context.l10n.defaultSearchHint,
      helper: context.l10n.defaultSearchHelper,
      fieldKey: const Key('default-search-field'),
      save: settings.setDefaultSearch,
    );
  }

  /// The platforms with the MCP server for agents.
  static const _desktop = {
    TargetPlatform.linux,
    TargetPlatform.macOS,
    TargetPlatform.windows,
  };

  static const _learnAheadLimits = [0, 5, 10, 15, 20, 30, 60, 120, 240, 1440];

  static String _learnAheadLabel(int minutes) {
    if (minutes == 0) return appL10n.off;
    if (minutes < 60 || minutes % 60 != 0) return appL10n.minutesShort(minutes);
    return appL10n.countHours(minutes ~/ 60);
  }

  static const _timeboxLimits = [0, 5, 10, 15, 20, 30, 45, 60, 90, 120];

  static String _timeboxLabel(int minutes) =>
      minutes == 0 ? appL10n.off : appL10n.minutesShort(minutes);

  static const _questionFades = [0, 50, 100, 200, 300, 500, 750, 1000];

  /// The UI sizes, 75 to 200 % in steps of 5 %.
  static final _uiScales = [for (var p = 75; p <= 200; p += 5) p];

  static String _percentLabel(int percent) =>
      NumberFormat.percentPattern(appL10n.localeName).format(percent / 100);

  static String _fadeLabel(int ms) =>
      ms == 0 ? appL10n.off : appL10n.millisecondsShort(ms);

  @override
  Widget build(BuildContext context) {
    final app = AppScope.of(context);
    final settings = app.settings;
    final l = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(l.settings)),
      body: ListenableBuilder(
        listenable: settings,
        builder: (context, _) => ListView(
          key: const Key('settings-list'),
          children: [
            if (_busy) const LinearProgressIndicator(),
            _Section(l.appearance),
            ListTile(
              title: Text(l.theme),
              subtitle: Padding(
                padding: const EdgeInsets.only(top: 8),
                // Scaled down where the tile is too narrow, as on a phone
                // with a large UI size.
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  alignment: AlignmentDirectional.centerStart,
                  child: SegmentedButton<ThemeMode>(
                    segments: [
                      ButtonSegment(
                        value: ThemeMode.system,
                        label: Text(l.themeSystem),
                        icon: const Icon(Icons.brightness_auto),
                      ),
                      ButtonSegment(
                        value: ThemeMode.light,
                        label: Text(l.themeLight),
                        icon: const Icon(Icons.light_mode),
                      ),
                      ButtonSegment(
                        value: ThemeMode.dark,
                        label: Text(l.themeDark),
                        icon: const Icon(Icons.dark_mode),
                      ),
                    ],
                    selected: {settings.themeMode},
                    onSelectionChanged: (s) => settings.setThemeMode(s.first),
                  ),
                ),
              ),
            ),
            _ChoiceTile(
              title: Text(l.language),
              control: DropdownButton<String>(
                key: const Key('language'),
                // A language without a translation shows as English, like
                // the system default does.
                value:
                    AppLocalizations.supportedLocales.any(
                      (locale) => locale.toLanguageTag() == settings.language,
                    )
                    ? settings.language
                    : '',
                items: [
                  DropdownMenuItem(value: '', child: Text(l.languageSystem)),
                  // Each language by its own name, so that it can be found
                  // whatever the current one.
                  for (final locale in AppLocalizations.supportedLocales)
                    DropdownMenuItem(
                      value: locale.toLanguageTag(),
                      child: Text(lookupAppLocalizations(locale).languageName),
                    ),
                ],
                onChanged: (code) {
                  if (code != null) settings.setLanguage(code);
                },
              ),
            ),
            _ChoiceTile(
              title: Text(l.uiSize),
              subtitle: Text(l.uiSizeHelp),
              control: DropdownButton<int>(
                key: const Key('ui-size'),
                value: settings.uiScale,
                items: [
                  for (final p in _with(_uiScales, settings.uiScale))
                    DropdownMenuItem(value: p, child: Text(_percentLabel(p))),
                ],
                onChanged: (p) {
                  if (p != null) settings.setUiScale(p);
                },
              ),
            ),
            SwitchListTile(
              key: const Key('minimalist'),
              title: Text(l.minimalistMode),
              subtitle: Text(l.minimalistModeHelp),
              value: settings.minimalist,
              onChanged: settings.setMinimalist,
            ),
            _ChoiceTile(
              title: Text(l.reduceMotion),
              subtitle: Text(l.reduceMotionHelp),
              control: DropdownButton<String>(
                key: const Key('reduce-motion'),
                value: settings.reduceMotion,
                items: [
                  DropdownMenuItem(
                    value: 'system',
                    child: Text(l.reduceMotionSystem),
                  ),
                  DropdownMenuItem(value: 'on', child: Text(l.on)),
                  DropdownMenuItem(value: 'off', child: Text(l.off)),
                ],
                onChanged: (mode) {
                  if (mode != null) settings.setReduceMotion(mode);
                },
              ),
            ),
            _Section(l.studying),
            _ChoiceTile(
              title: Text(l.nextDayStarts),
              subtitle: Text(l.nextDayStartsHelp),
              control: DropdownButton<int>(
                key: const Key('rollover'),
                value: settings.rolloverHour,
                items: [
                  for (var h = 0; h < 24; h++)
                    DropdownMenuItem(
                      value: h,
                      child: Text('${h.toString().padLeft(2, '0')}:00'),
                    ),
                ],
                onChanged: (h) {
                  if (h != null) settings.setRolloverHour(h);
                },
              ),
            ),
            _ChoiceTile(
              title: Text(l.learnAheadLimit),
              subtitle: Text(l.learnAheadLimitHelp),
              control: DropdownButton<int>(
                key: const Key('learnAhead'),
                value: settings.learnAheadMinutes,
                items: [
                  for (final m in _with(
                    _learnAheadLimits,
                    settings.learnAheadMinutes,
                  ))
                    DropdownMenuItem(
                      value: m,
                      child: Text(_learnAheadLabel(m)),
                    ),
                ],
                onChanged: (m) {
                  if (m != null) settings.setLearnAheadMinutes(m);
                },
              ),
            ),
            _ChoiceTile(
              title: Text(l.timeboxLimit),
              subtitle: Text(l.timeboxLimitHelp),
              control: DropdownButton<int>(
                key: const Key('timebox'),
                value: settings.timeboxMinutes,
                items: [
                  for (final m in _with(
                    _timeboxLimits,
                    settings.timeboxMinutes,
                  ))
                    DropdownMenuItem(value: m, child: Text(_timeboxLabel(m))),
                ],
                onChanged: (m) {
                  if (m != null) settings.setTimeboxMinutes(m);
                },
              ),
            ),
            SwitchListTile(
              title: Text(l.autoContinueLearning),
              subtitle: Text(l.autoContinueLearningHelp),
              value: settings.autoContinueLearning,
              onChanged: settings.setAutoContinueLearning,
            ),
            SwitchListTile(
              title: Text(l.showPlayButtons),
              value: settings.showPlayButtons,
              onChanged: settings.setShowPlayButtons,
            ),
            SwitchListTile(
              title: Text(l.interruptAudio),
              subtitle: Text(l.interruptAudioHelp),
              value: settings.interruptAudio,
              onChanged: settings.setInterruptAudio,
            ),
            SwitchListTile(
              key: const Key('strip-paste-formatting'),
              title: Text(l.stripPasteFormatting),
              subtitle: Text(l.stripPasteFormattingHelp(commandKey)),
              value: settings.stripPasteFormatting,
              onChanged: settings.setStripPasteFormatting,
            ),
            SwitchListTile(
              key: const Key('paste-images-as-png'),
              title: Text(l.pasteImagesAsPng),
              subtitle: Text(l.pasteImagesAsPngHelp),
              value: settings.pasteImagesAsPng,
              onChanged: settings.setPasteImagesAsPng,
            ),
            ListTile(
              title: Text(l.deckForNewCards),
              subtitle: DropdownButton<bool>(
                key: const Key('deck-for-new-cards'),
                isExpanded: true,
                value: settings.deckByNoteType,
                items: [
                  DropdownMenuItem(value: false, child: Text(l.currentDeck)),
                  DropdownMenuItem(value: true, child: Text(l.deckByNoteType)),
                ],
                onChanged: (on) {
                  if (on != null) settings.setDeckByNoteType(on);
                },
              ),
            ),
            _Section(l.goalsAndStreaks),
            SwitchListTile(
              key: const Key('show-progress'),
              title: Text(l.showStreaksAndGoals),
              subtitle: Text(l.showStreaksAndGoalsHelp),
              value: settings.showProgress,
              onChanged: settings.setShowProgress,
            ),
            SwitchListTile(
              key: const Key('deck-progress-in-list'),
              title: Text(l.deckProgressInList),
              subtitle: Text(l.deckProgressInListHelp),
              value: settings.deckProgressInList,
              onChanged: settings.setDeckProgressInList,
            ),
            ListTile(
              key: const Key('daily-goal'),
              title: Text(l.dailyGoal),
              subtitle: Text(
                dailyGoalLabel(l, settings.dailyGoalKind, settings.dailyGoal),
              ),
              onTap: () => editDailyGoal(context),
            ),
            _ChoiceTile(
              title: Text(l.restDaysPerWeek),
              subtitle: Text(l.restDaysPerWeekHelp),
              control: DropdownButton<int>(
                key: const Key('rest-days'),
                value: settings.restDaysPerWeek,
                items: [
                  for (var n = 0; n <= maxRestDaysPerWeek; n++)
                    DropdownMenuItem(value: n, child: Text('$n')),
                ],
                onChanged: (n) {
                  if (n != null) settings.setRestDaysPerWeek(n);
                },
              ),
            ),
            _Section(l.review),
            SwitchListTile(
              key: const Key('show-remaining-counts'),
              title: Text(l.showRemainingCounts),
              value: settings.showRemainingCounts,
              onChanged: settings.setShowRemainingCounts,
            ),
            SwitchListTile(
              key: const Key('show-intervals'),
              title: Text(l.showIntervals),
              value: settings.showIntervals,
              onChanged: settings.setShowIntervals,
            ),
            SwitchListTile(
              key: const Key('space-answers-good'),
              title: Text(l.spaceAnswersGood),
              subtitle: Text(l.spaceAnswersGoodHelp),
              value: settings.spaceAnswersGood,
              onChanged: settings.setSpaceAnswersGood,
            ),
            SwitchListTile(
              key: const Key('hide-bars-during-review'),
              title: Text(l.hideBarsDuringReview),
              subtitle: Text(l.hideBarsDuringReviewHelp),
              value: settings.hideBarsDuringReview,
              onChanged: settings.setHideBarsDuringReview,
            ),
            _ChoiceTile(
              title: Text(l.questionFade),
              subtitle: Text(l.questionFadeHelp),
              control: DropdownButton<int>(
                key: const Key('question-fade'),
                value: settings.questionFadeMs,
                items: [
                  for (final ms in _with(
                    _questionFades,
                    settings.questionFadeMs,
                  ))
                    DropdownMenuItem(value: ms, child: Text(_fadeLabel(ms))),
                ],
                onChanged: (ms) {
                  if (ms != null) settings.setQuestionFadeMs(ms);
                },
              ),
            ),
            _Section(l.browsing),
            SwitchListTile(
              key: const Key('ignore-accents'),
              title: Text(l.ignoreAccents),
              subtitle: Text(l.ignoreAccentsHelp),
              value: settings.ignoreAccents,
              onChanged: settings.setIgnoreAccents,
            ),
            ListTile(
              key: const Key('default-search'),
              title: Text(l.defaultBrowserSearch),
              subtitle: Text(
                settings.defaultSearch.isEmpty
                    ? l.allCards
                    : settings.defaultSearch,
              ),
              onTap: _editDefaultSearch,
            ),
            _Section(l.profile),
            ListenableBuilder(
              listenable: app.profiles,
              builder: (context, _) => Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  ListTile(
                    key: const Key('switch-profile'),
                    leading: const Icon(Icons.switch_account_outlined),
                    title: Text(l.switchProfile),
                    subtitle: Text(app.profiles.currentName ?? ''),
                    onTap: () => Navigator.of(context).push(
                      MaterialPageRoute<void>(
                        builder: (_) => ProfileListScreen(
                          profiles: app.profiles,
                          onOpen: app.switchProfile,
                        ),
                      ),
                    ),
                  ),
                  SwitchListTile(
                    key: const Key('profiles-at-startup'),
                    title: Text(l.showProfilesAtStartup),
                    subtitle: Text(l.showProfilesAtStartupHelp),
                    value: app.profiles.list.showAtStartup,
                    onChanged: (on) =>
                        guard(() => app.profiles.setShowAtStartup(on)),
                  ),
                ],
              ),
            ),
            if (!app.isWeb) SyncSection(heading: _Section(l.sync)),
            _Section(l.collection),
            ListTile(
              leading: const Icon(Icons.file_open_outlined),
              title: Text(l.import),
              subtitle: Text(l.importHelp),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute<void>(builder: (_) => const ImportScreen()),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.save_alt),
              title: Text(l.exportBackup),
              subtitle: Text(
                app.isWeb ? l.downloadsBackupFile : l.savesBackupFile,
              ),
              enabled: !_busy,
              onTap: _export,
            ),
            ListTile(
              leading: const Icon(Icons.ios_share),
              title: Text(l.exportCollection),
              subtitle: Text(l.exportCollectionHelp),
              enabled: !_busy,
              onTap: () => showExportDialog(context, const CollectionExport()),
            ),
            ListTile(
              leading: const Icon(Icons.restore),
              title: Text(l.restoreBackup),
              subtitle: Text(l.restoreBackupHelp),
              enabled: !_busy,
              onTap: _restore,
            ),
            ListTile(
              leading: const Icon(Icons.cleaning_services_outlined),
              title: Text(l.emptyCards),
              subtitle: Text(l.emptyCardsHelp),
              onTap: () => showDialog<void>(
                context: context,
                animationStyle: popupAnimation(context),
                builder: (_) => const _EmptyCardsDialog(),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.perm_media_outlined),
              title: Text(l.checkMedia),
              subtitle: Text(l.checkMediaHelp),
              onTap: () => showDialog<void>(
                context: context,
                animationStyle: popupAnimation(context),
                builder: (_) => const _CheckMediaDialog(),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.health_and_safety_outlined),
              title: Text(l.checkDatabase),
              subtitle: Text(l.checkDatabaseHelp),
              onTap: () => showDialog<void>(
                context: context,
                animationStyle: popupAnimation(context),
                barrierDismissible: false,
                builder: (_) => const _CheckDatabaseDialog(),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.history),
              title: Text(l.exportDeletionLog),
              subtitle: Text(l.exportDeletionLogHelp),
              enabled: !_busy,
              onTap: _exportDeletionLog,
            ),
            ListTile(
              leading: const Icon(Icons.delete_sweep_outlined),
              title: Text(l.clearDeletionLog),
              subtitle: Text(l.clearDeletionLogHelp),
              enabled: !_busy,
              onTap: _clearDeletionLog,
            ),
            _Section(l.backups),
            const _BackupSection(),
            _Section(l.flags),
            for (var flag = 1; flag <= flagCount; flag++)
              ListTile(
                key: Key('flag$flag'),
                leading: Icon(
                  Icons.flag,
                  color: FlagColors.of(context).flag(flag),
                ),
                title: Text(settings.flagName(flag)),
                subtitle: settings.flagName(flag) == flagColorName(flag, l)
                    ? null
                    : Text(flagColorName(flag, l)),
                trailing: const Icon(Icons.edit_outlined),
                onTap: () => _renameFlag(flag),
              ),
            if (!app.isWeb && _desktop.contains(defaultTargetPlatform))
              const _McpSection(),
            FutureBuilder<golang.Status>(
              future: _status,
              builder: (context, snap) {
                final status = snap.data;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (app.isWeb && status != null)
                      ListTile(
                        leading: Icon(
                          status.persistent
                              ? Icons.verified_outlined
                              : Icons.warning_amber,
                        ),
                        title: Text(l.storage),
                        subtitle: Text(
                          status.persistent
                              ? l.storagePersistent
                              : l.storageNotPersistent,
                        ),
                      ),
                    _Section(l.about),
                    ListTile(
                      leading: const Icon(Icons.style_outlined),
                      title: const Text('Cardflip'),
                      subtitle: Text(l.aboutCardflip(status?.version ?? '…')),
                      isThreeLine: true,
                    ),
                    ListTile(
                      leading: const Icon(Icons.description_outlined),
                      title: Text(l.openSourceLicenses),
                      onTap: () => showLicensePage(
                        context: context,
                        applicationName: 'Cardflip',
                        applicationVersion: status?.version,
                        applicationIcon: Image.asset(
                          'assets/logo/cardflip_logo_1024.png',
                          width: 64,
                          height: 64,
                        ),
                        applicationLegalese: l.licenseLegalese,
                      ),
                    ),
                    ListTile(
                      leading: const Icon(Icons.code),
                      title: Text(l.sourceCode),
                      trailing: const Icon(Icons.open_in_new),
                      onTap: () => guard(
                        () => SettingsScreen.openUrl(Uri.parse(repositoryUrl)),
                      ),
                    ),
                  ],
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

/// Backup settings, "create backup now" and the automatic backups, which
/// can be restored.
class _BackupSection extends StatefulWidget {
  const _BackupSection();

  @override
  State<_BackupSection> createState() => _BackupSectionState();
}

class _BackupSectionState extends State<_BackupSection> {
  /// How many backups are listed until "Show all" is tapped.
  static const _collapsedCount = 5;
  static const _intervals = [0, 5, 10, 15, 30, 60, 120, 240, 480, 1440];
  static const _counts = [
    0,
    1,
    2,
    3,
    4,
    5,
    6,
    7,
    8,
    10,
    12,
    15,
    20,
    30,
    50,
    100,
  ];

  bool _showAll = false;

  static String _intervalLabel(int minutes) {
    final l = appL10n;
    if (minutes == 0) return l.off;
    if (minutes < 60 || minutes % 60 != 0) return l.backupEveryMinutes(minutes);
    if (minutes == 60) return l.backupEveryHour;
    return l.backupEveryHours(minutes ~/ 60);
  }

  Future<void> _restore(golang.AutoBackup backup) async {
    final ok = await confirm(
      context,
      title: context.l10n.replaceCollectionTitle,
      message: context.l10n.restoreAutoBackupBody(
        formatDateTime(backup.createdAt),
      ),
      action: context.l10n.restore,
      destructive: true,
    );
    if (!ok || !mounted) return;
    await AppScope.read(context).backups.restore(backup.name);
  }

  Widget _count(
    String key,
    String label,
    int value,
    golang.BackupSettings Function(int) update,
  ) {
    final settings = AppScope.of(context).settings;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(label),
        const SizedBox(width: 8),
        DropdownButton<int>(
          key: Key('keep$key'),
          value: value,
          items: [
            for (final n in _with(_counts, value))
              DropdownMenuItem(value: n, child: Text('$n')),
          ],
          onChanged: (n) {
            if (n != null) settings.setBackupSettings(update(n));
          },
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final app = AppScope.of(context);
    final settings = app.settings;
    return ListenableBuilder(
      listenable: Listenable.merge([app.backups, settings]),
      builder: (context, _) {
        final model = app.backups;
        final b = settings.backup;
        final list = model.backups;
        final shown = _showAll ? list : list.take(_collapsedCount).toList();
        final l = context.l10n;
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _ChoiceTile(
              title: Text(l.backUpAutomatically),
              subtitle: Text(l.backUpAutomaticallyHelp),
              control: DropdownButton<int>(
                key: const Key('backupInterval'),
                value: b.intervalMinutes,
                items: [
                  for (final m in _with(_intervals, b.intervalMinutes))
                    DropdownMenuItem(value: m, child: Text(_intervalLabel(m))),
                ],
                onChanged: (m) {
                  if (m == null) return;
                  settings.setBackupSettings(
                    golang.BackupSettings(
                      intervalMinutes: m,
                      daily: b.daily,
                      weekly: b.weekly,
                      monthly: b.monthly,
                    ),
                  );
                },
              ),
            ),
            ListTile(
              title: Text(l.keepBackups),
              subtitle: Wrap(
                spacing: 24,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  _count(
                    'Daily',
                    l.keepDaily,
                    b.daily,
                    (n) => golang.BackupSettings(
                      intervalMinutes: b.intervalMinutes,
                      daily: n,
                      weekly: b.weekly,
                      monthly: b.monthly,
                    ),
                  ),
                  _count(
                    'Weekly',
                    l.keepWeekly,
                    b.weekly,
                    (n) => golang.BackupSettings(
                      intervalMinutes: b.intervalMinutes,
                      daily: b.daily,
                      weekly: n,
                      monthly: b.monthly,
                    ),
                  ),
                  _count(
                    'Monthly',
                    l.keepMonthly,
                    b.monthly,
                    (n) => golang.BackupSettings(
                      intervalMinutes: b.intervalMinutes,
                      daily: b.daily,
                      weekly: b.weekly,
                      monthly: n,
                    ),
                  ),
                ],
              ),
            ),
            ListTile(
              leading: const Icon(Icons.backup_outlined),
              title: Text(l.createBackupNow),
              subtitle: showShortcutHints ? Text(backupShortcutLabel) : null,
              enabled: !model.busy,
              onTap: model.create,
            ),
            ListTile(
              key: const Key('backupSpace'),
              leading: const Icon(Icons.storage_outlined),
              title: Text(
                list.isEmpty
                    ? l.noAutoBackups
                    : l.countAutoBackups(list.length),
              ),
              subtitle: Text(l.usedOnDevice(formatBytes(model.bytes))),
            ),
            for (final backup in shown)
              ListTile(
                key: ValueKey(backup.name),
                contentPadding: const EdgeInsetsDirectional.only(
                  start: 72,
                  end: 16,
                ),
                title: Text(formatDateTime(backup.createdAt)),
                subtitle: Text(formatBytes(backup.size)),
                trailing: TextButton(
                  onPressed: model.busy ? null : () => _restore(backup),
                  child: Text(l.restore),
                ),
              ),
            if (list.length > _collapsedCount)
              Align(
                alignment: AlignmentDirectional.centerStart,
                child: Padding(
                  padding: const EdgeInsetsDirectional.only(start: 60),
                  child: TextButton(
                    onPressed: () => setState(() => _showAll = !_showAll),
                    child: Text(
                      _showAll ? l.showFewer : l.showAllCount(list.length),
                    ),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

/// The MCP server for agents: the switch, and while it is on, its status,
/// port, token, a client configuration to copy and the agents' recent
/// changes. Shown on the desktop platforms only, where Go has the server.
class _McpSection extends StatefulWidget {
  const _McpSection();

  @override
  State<_McpSection> createState() => _McpSectionState();
}

class _McpSectionState extends State<_McpSection> {
  bool _supported = false;
  golang.McpStatus? _status;
  List<golang.McpActivity> _activity = const [];
  StreamSubscription<golang.McpStatus>? _changes;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final supported = await golang.mcp.supported().catchError((Object _) {
      return false;
    });
    if (!supported || !mounted) return;
    setState(() => _supported = true);
    _changes = golang.mcp.statusChanges().listen((status) async {
      final activity = await golang.mcp.activity();
      if (mounted) {
        setState(() {
          _status = status;
          _activity = activity;
        });
      }
    }, onError: (Object e) => showError(e));
  }

  @override
  void dispose() {
    _changes?.cancel();
    super.dispose();
  }

  Future<void> _update(Future<golang.McpStatus> Function() call) async {
    await guard(() async {
      final status = await call();
      if (mounted) setState(() => _status = status);
    });
  }

  Future<void> _setEnabled(bool on) async {
    // The token is created when the server is first turned on.
    if (on && _status!.token.isEmpty) {
      final ok = await confirm(
        context,
        title: context.l10n.allowAgentsTitle,
        message: context.l10n.allowAgentsBody,
        action: context.l10n.allow,
      );
      if (!ok) return;
    }
    await _update(() => golang.mcp.setEnabled(on));
  }

  Future<void> _editPort() async {
    final port = await askText(
      context,
      title: context.l10n.mcpPort,
      initial: '${_status!.port}',
      label: context.l10n.mcpPort,
      helper: context.l10n.mcpPortHelper,
    );
    if (port == null) return;
    // Go explains a port that is not a number or out of range.
    await _update(() => golang.mcp.setPort(int.tryParse(port.trim()) ?? 0));
  }

  Future<void> _regenerate() async {
    final ok = await confirm(
      context,
      title: context.l10n.regenerateTokenTitle,
      message: context.l10n.regenerateTokenBody,
      action: context.l10n.regenerate,
      destructive: true,
    );
    if (ok) await _update(golang.mcp.regenerateToken);
  }

  Future<void> _copy(String text, String done) async {
    await Clipboard.setData(ClipboardData(text: text));
    showInfo(done);
  }

  /// The configuration MCP clients such as Claude Code read.
  static String clientConfig(golang.McpStatus s) =>
      const JsonEncoder.withIndent('  ').convert({
        'mcpServers': {
          'cardflip': {
            'type': 'http',
            'url': s.url,
            'headers': {'Authorization': 'Bearer ${s.token}'},
          },
        },
      });

  @override
  Widget build(BuildContext context) {
    final status = _status;
    if (!_supported || status == null) return const SizedBox.shrink();
    final l = context.l10n;
    final error = status.error;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _Section(l.agentsMcp),
        SwitchListTile(
          key: const Key('mcp-switch'),
          secondary: const Icon(Icons.smart_toy_outlined),
          title: Text(l.allowAgents),
          subtitle: Text(l.allowAgentsHelp),
          value: status.enabled,
          onChanged: _setEnabled,
        ),
        if (status.enabled) ...[
          ListTile(
            key: const Key('mcp-status'),
            leading: Icon(
              status.running ? Icons.check_circle_outline : Icons.error_outline,
              color: status.running
                  ? null
                  : Theme.of(context).colorScheme.error,
            ),
            title: Text(
              status.running ? l.mcpServerRunning : l.mcpServerStopped,
            ),
            subtitle: SelectableText(
              status.running || error == null ? status.url : goText(error),
            ),
          ),
          ListTile(
            key: const Key('mcp-port'),
            leading: const Icon(Icons.settings_ethernet),
            title: Text(l.mcpPort),
            subtitle: Text('${status.port}'),
            trailing: const Icon(Icons.edit_outlined),
            onTap: _editPort,
          ),
          ListTile(
            key: const Key('mcp-token'),
            leading: const Icon(Icons.key_outlined),
            title: Text(l.mcpToken),
            subtitle: Text('•' * 16),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  key: const Key('mcp-copy-token'),
                  tooltip: l.copyToken,
                  icon: const Icon(Icons.copy),
                  onPressed: () => _copy(status.token, l.tokenCopied),
                ),
                IconButton(
                  key: const Key('mcp-regenerate-token'),
                  tooltip: l.regenerateToken,
                  icon: const Icon(Icons.refresh),
                  onPressed: _regenerate,
                ),
              ],
            ),
          ),
          ListTile(
            key: const Key('mcp-copy-config'),
            leading: const Icon(Icons.content_copy),
            title: Text(l.copyClientConfig),
            subtitle: Text(l.copyClientConfigHelp),
            onTap: () => _copy(clientConfig(status), l.clientConfigCopied),
          ),
          ListTile(
            leading: const Icon(Icons.history),
            title: Text(l.agentActivity),
            subtitle: _activity.isEmpty ? Text(l.noAgentActivity) : null,
          ),
          for (final (i, a) in _activity.indexed)
            ListTile(
              key: Key('mcp-activity-$i'),
              dense: true,
              contentPadding: const EdgeInsetsDirectional.only(
                start: 72,
                end: 16,
              ),
              title: Text(goText(a.step)),
              subtitle: Text(formatDateTime(a.time.toLocal())),
            ),
        ],
      ],
    );
  }
}

/// Lists the empty cards by note and deletes them when confirmed. Notes
/// whose cards are all empty keep one card.
class _EmptyCardsDialog extends StatefulWidget {
  const _EmptyCardsDialog();

  @override
  State<_EmptyCardsDialog> createState() => _EmptyCardsDialogState();
}

class _EmptyCardsDialogState extends State<_EmptyCardsDialog> {
  late final Future<golang.EmptyCardList> _list = golang.app.emptyCards();
  bool _busy = false;

  Future<void> _delete(golang.EmptyCardList list) async {
    setState(() => _busy = true);
    final ok = await guard(
      () => golang.app.deleteEmptyCards([
        for (final n in list.notes)
          for (final c in n.cards) c.cardID,
      ]),
    );
    if (!mounted) return;
    if (!ok) {
      setState(() => _busy = false);
      return;
    }
    Navigator.pop(context);
    showInfo(appL10n.deletedEmptyCards(list.deletable));
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<golang.EmptyCardList>(
      future: _list,
      builder: (context, snap) {
        final list = snap.data;
        final l = context.l10n;
        final close = TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(list != null && list.total > 0 ? l.cancel : l.close),
        );
        if (snap.hasError) {
          return AlertDialog(
            title: Text(l.emptyCards),
            content: Text(errorMessage(snap.error!)),
            actions: [close],
          );
        }
        if (list == null) {
          return AlertDialog(
            title: Text(l.emptyCards),
            content: const SizedBox(
              height: 64,
              child: Center(child: CircularProgressIndicator()),
            ),
          );
        }
        if (list.total == 0) {
          return AlertDialog(
            title: Text(l.emptyCards),
            content: Text(l.noEmptyCards),
            actions: [close],
          );
        }
        final kept = list.notes.where((n) => n.allEmpty).length;
        final scheme = Theme.of(context).colorScheme;
        return AlertDialog(
          title: Text(l.emptyCards),
          content: SizedBox(
            width: 480,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(l.emptyCardsSummary(list.total, list.notes.length)),
                if (kept > 0) ...[
                  const SizedBox(height: 8),
                  Text(l.emptyCardsKept(kept)),
                ],
                const SizedBox(height: 8),
                Flexible(
                  child: ListView.builder(
                    shrinkWrap: true,
                    itemCount: list.notes.length,
                    itemBuilder: (context, i) {
                      final n = list.notes[i];
                      return ListTile(
                        key: ValueKey(n.noteID),
                        dense: true,
                        contentPadding: EdgeInsets.zero,
                        title: Text(
                          n.text.isEmpty ? l.emptyText : n.text,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        subtitle: Text(
                          [
                            n.cards.map((c) => c.name).join(', '),
                            if (n.allEmpty) l.oneCardKept,
                          ].join(' · '),
                          style: n.allEmpty
                              ? TextStyle(color: scheme.tertiary)
                              : null,
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
          actions: [
            close,
            FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: scheme.error,
                foregroundColor: scheme.onError,
              ),
              onPressed: _busy || list.deletable == 0
                  ? null
                  : () => _delete(list),
              child: Text(l.deleteCards(list.deletable)),
            ),
          ],
        );
      },
    );
  }
}

/// The media report: files nothing refers to, files notes refer to that
/// are missing, and the media trash, with their actions.
class _CheckMediaDialog extends StatefulWidget {
  const _CheckMediaDialog();

  @override
  State<_CheckMediaDialog> createState() => _CheckMediaDialogState();
}

class _CheckMediaDialogState extends State<_CheckMediaDialog> {
  late Future<golang.MediaCheck> _report = golang.app.checkMedia();
  bool _busy = false;

  /// Runs [action] after the user confirmed, then checks again.
  Future<void> _act({
    required String title,
    required String message,
    required String action,
    bool destructive = false,
    required Future<void> Function() run,
  }) async {
    final ok = await confirm(
      context,
      title: title,
      message: message,
      action: action,
      destructive: destructive,
    );
    if (!ok || !mounted) return;
    setState(() => _busy = true);
    await guard(run);
    if (!mounted) return;
    setState(() {
      _busy = false;
      _report = golang.app.checkMedia();
    });
  }

  static String _files(List<golang.MediaFile> files, int bytes) =>
      appL10n.filesWithSize(files.length, formatBytes(bytes));

  /// Drops cached media, so that cards show the current files.
  static void _forgetMedia() {
    MediaCache.instance.clear();
    FontCache.instance.clear();
    StoredMediaFiles.forget();
  }

  Future<void> _deleteUnused(golang.MediaCheck r) => _act(
    title: appL10n.deleteUnusedTitle,
    message: appL10n.deleteUnusedBody(_files(r.unused, r.unusedBytes)),
    action: appL10n.delete,
    destructive: true,
    run: () async {
      final n = await golang.app.deleteUnusedMedia();
      _forgetMedia();
      showInfo(appL10n.movedToTrash(n));
    },
  );

  Future<void> _tagMissing(golang.MediaCheck r) => _act(
    title: appL10n.tagMissingTitle,
    message: appL10n.tagMissingBody(r.missingNotes),
    action: appL10n.tag,
    run: () async {
      final n = await golang.app.tagMissingMedia();
      showInfo(appL10n.taggedMissing(n));
    },
  );

  Future<void> _restore(golang.MediaCheck r) => _act(
    title: appL10n.restoreDeletedTitle,
    message: appL10n.restoreDeletedBody(_files(r.trash, r.trashBytes)),
    action: appL10n.restore,
    run: () async {
      final n = await golang.app.restoreMediaTrash();
      _forgetMedia();
      final kept = r.trash.length - n;
      showInfo(
        kept > 0
            ? appL10n.restoredFilesKept(n, kept)
            : appL10n.restoredFiles(n),
      );
    },
  );

  Future<void> _emptyTrash(golang.MediaCheck r) => _act(
    title: appL10n.emptyTrashTitle,
    message: appL10n.emptyTrashBody(_files(r.trash, r.trashBytes)),
    action: appL10n.emptyTrash,
    destructive: true,
    run: () async {
      await golang.app.emptyMediaTrash();
      showInfo(appL10n.trashEmptied);
    },
  );

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final close = TextButton(
      onPressed: () => Navigator.pop(context),
      child: Text(l.close),
    );
    return FutureBuilder<golang.MediaCheck>(
      future: _report,
      builder: (context, snap) {
        if (snap.hasError) {
          return AlertDialog(
            title: Text(l.checkMedia),
            content: Text(errorMessage(snap.error!)),
            actions: [close],
          );
        }
        final r = snap.data;
        if (r == null || snap.connectionState != ConnectionState.done) {
          return AlertDialog(
            title: Text(l.checkMedia),
            content: const SizedBox(
              height: 64,
              child: Center(child: CircularProgressIndicator()),
            ),
          );
        }
        final theme = Theme.of(context);
        final muted = TextStyle(color: theme.colorScheme.onSurfaceVariant);
        Widget heading(String text) => Padding(
          padding: const EdgeInsets.only(top: 16, bottom: 4),
          child: Text(
            text,
            style: theme.textTheme.titleSmall?.copyWith(
              color: theme.colorScheme.primary,
            ),
          ),
        );
        Widget file(String name, [int? size]) => ListTile(
          key: ValueKey(name),
          dense: true,
          visualDensity: VisualDensity.compact,
          contentPadding: EdgeInsets.zero,
          title: Text(name, maxLines: 1, overflow: TextOverflow.ellipsis),
          trailing: size == null ? null : Text(formatBytes(size), style: muted),
        );
        Widget buttons(List<Widget> children) => Padding(
          padding: const EdgeInsets.only(top: 4),
          child: Wrap(spacing: 8, runSpacing: 4, children: children),
        );
        return AlertDialog(
          title: Text(l.checkMedia),
          content: SizedBox(
            width: 480,
            child: ListView(
              shrinkWrap: true,
              children: [
                heading(l.unusedFiles),
                Text(
                  r.unused.isEmpty
                      ? l.allMediaUsed
                      : l.unusedFilesBody(_files(r.unused, r.unusedBytes)),
                ),
                for (final f in r.unused) file(f.name, f.size),
                if (r.unused.isNotEmpty)
                  buttons([
                    OutlinedButton.icon(
                      icon: const Icon(Icons.delete_outline),
                      label: Text(l.deleteUnused),
                      onPressed: _busy ? null : () => _deleteUnused(r),
                    ),
                  ]),
                heading(l.missingFiles),
                Text(
                  r.missing.isEmpty
                      ? l.noMissingFiles
                      : l.missingFilesBody(r.missing.length, r.missingNotes),
                ),
                for (final name in r.missing) file(name),
                if (r.missing.isNotEmpty)
                  buttons([
                    OutlinedButton.icon(
                      icon: const Icon(Icons.sell_outlined),
                      label: Text(l.tagMissing),
                      onPressed: _busy ? null : () => _tagMissing(r),
                    ),
                  ]),
                heading(l.mediaTrash),
                Text(
                  r.trash.isEmpty
                      ? l.trashIsEmpty
                      : l.trashBody(_files(r.trash, r.trashBytes)),
                ),
                if (r.trash.isNotEmpty)
                  buttons([
                    OutlinedButton.icon(
                      icon: const Icon(Icons.restore_from_trash_outlined),
                      label: Text(l.restoreDeleted),
                      onPressed: _busy ? null : () => _restore(r),
                    ),
                    OutlinedButton.icon(
                      icon: const Icon(Icons.delete_forever_outlined),
                      label: Text(l.emptyTrash),
                      onPressed: _busy ? null : () => _emptyTrash(r),
                    ),
                  ]),
              ],
            ),
          ),
          actions: [close],
        );
      },
    );
  }
}

/// Checks the collection and repairs what it can, showing progress and
/// then the report. It cannot be dismissed while the check runs.
class _CheckDatabaseDialog extends StatefulWidget {
  const _CheckDatabaseDialog();

  @override
  State<_CheckDatabaseDialog> createState() => _CheckDatabaseDialogState();
}

class _CheckDatabaseDialogState extends State<_CheckDatabaseDialog> {
  late final StreamSubscription<golang.CheckDatabaseEvent> _check;
  golang.CheckDatabaseEvent? _progress;
  golang.CheckReport? _report;
  String? _error;

  @override
  void initState() {
    super.initState();
    _check = golang.app.checkDatabase().listen(
      (e) {
        if (!mounted) return;
        setState(() {
          if (e.error.isNotEmpty) {
            _error = errorMessage(
              golang.GolangException('app.CheckDatabase', e.error),
            );
          } else if (e.report case final r?) {
            _report = r;
          } else {
            _progress = e;
          }
        });
      },
      onError: (Object e) {
        if (mounted) setState(() => _error = errorMessage(e));
      },
    );
  }

  @override
  void dispose() {
    _check.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final title = Text(l.checkDatabase);
    final close = TextButton(
      onPressed: () => Navigator.pop(context),
      child: Text(l.close),
    );
    if (_error case final error?) {
      return AlertDialog(
        title: title,
        content: Text(error, key: const Key('check-error')),
        actions: [close],
      );
    }
    final r = _report;
    if (r == null) {
      final p = _progress;
      final stage = switch (p?.stage) {
        'notes' => l.checkingNotes,
        'cards' => l.checkingCards,
        'history' => l.readingHistory,
        'saving' => l.savingRepairs,
        _ => l.backingUp,
      };
      final value = p == null || p.total == 0 ? null : p.done / p.total;
      return AlertDialog(
        title: title,
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(stage, key: const Key('check-stage')),
            const SizedBox(height: 16),
            LinearProgressIndicator(
              key: const Key('check-progress'),
              value: value,
            ),
          ],
        ),
      );
    }
    final muted = TextStyle(
      color: Theme.of(context).colorScheme.onSurfaceVariant,
    );
    return AlertDialog(
      title: title,
      content: SizedBox(
        width: 480,
        child: ListView(
          key: const Key('check-report'),
          shrinkWrap: true,
          children: [
            if (r.problems.isEmpty)
              Text(l.noProblemsFound)
            else
              for (final p in r.problems)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Text(goText(p.message)),
                ),
            if (r.orphanedReviews > 0)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(l.orphanedReviews(r.orphanedReviews), style: muted),
              ),
          ],
        ),
      ),
      actions: [close],
    );
  }
}

/// [choices] plus [current], when it was set to something else.
List<int> _with(List<int> choices, int current) =>
    choices.contains(current) ? choices : ([...choices, current]..sort());

/// A setting with a [control], such as a dropdown, beside its title; below
/// the title where the tile is too narrow for both, as on a phone with a
/// large UI size.
class _ChoiceTile extends StatelessWidget {
  const _ChoiceTile({
    required this.title,
    this.subtitle,
    required this.control,
  });

  final Widget title;
  final Widget? subtitle;
  final Widget control;

  /// Narrower tiles show the control below the title.
  static const _minWidth = 360.0;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, constraints) {
      if (constraints.maxWidth >= _minWidth) {
        return ListTile(title: title, subtitle: subtitle, trailing: control);
      }
      return ListTile(
        title: title,
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ?subtitle,
            FittedBox(
              fit: BoxFit.scaleDown,
              alignment: AlignmentDirectional.centerStart,
              child: control,
            ),
          ],
        ),
      );
    },
  );
}

class _Section extends StatelessWidget {
  const _Section(this.title);

  final String title;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 24, 16, 8),
      child: Text(
        title,
        style: Theme.of(context).textTheme.titleSmall
            ?.copyWith(color: Theme.of(context).colorScheme.primary),
      ),
    );
  }
}
