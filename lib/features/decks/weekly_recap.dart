import 'package:flutter/material.dart';

import 'package:cardflip/data/deck_list_model.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/features/stats/stats_screen.dart';
import 'package:cardflip/features/stats/stats_sections.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// The deck list's weekly recap (progress spec, "Weekly recap"): last
/// week's study days, answers, study time and true retention, each with its
/// change from the week before. Weeks start on the first day of the week
/// of the interface language. Tapping it opens the collection statistics;
/// dismissing it hides it on this device until the next week. It is hidden
/// while streaks and goals are turned off, and when last week had no study.
///
/// It asks Go again whenever the deck list reloads, which notices the new
/// week; Go keeps the recap until the week or the review log changes.
class WeeklyRecap extends StatefulWidget {
  const WeeklyRecap({super.key});

  @override
  State<WeeklyRecap> createState() => _WeeklyRecapState();
}

class _WeeklyRecapState extends State<WeeklyRecap> {
  DeckListModel? _deckList;
  SettingsModel? _settings;
  int? _firstWeekday;
  golang.Recap? _recap;
  bool _shown = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final app = AppScope.of(context);
    final firstWeekday = MaterialLocalizations.of(context).firstDayOfWeekIndex;
    var changed = firstWeekday != _firstWeekday;
    _firstWeekday = firstWeekday;
    if (!identical(app.deckList, _deckList)) {
      _deckList?.removeListener(_load);
      _deckList = app.deckList..addListener(_load);
      changed = true;
    }
    if (!identical(app.settings, _settings)) {
      _settings?.removeListener(_settingsChanged);
      _settings = app.settings..addListener(_settingsChanged);
      changed = true;
    }
    if (changed) _load();
  }

  @override
  void dispose() {
    _deckList?.removeListener(_load);
    _settings?.removeListener(_settingsChanged);
    super.dispose();
  }

  /// Loads when streaks and goals are shown again, and shows or hides the
  /// recap when it is dismissed.
  void _settingsChanged() {
    final shown = _settings!.showProgress;
    if (shown && !_shown) {
      _load();
    } else {
      _shown = shown;
      setState(() {});
    }
  }

  Future<void> _load() async {
    _shown = _settings!.showProgress;
    if (!_shown) return;
    try {
      final recap = await golang.progress.weeklyRecap(_firstWeekday!);
      if (mounted) setState(() => _recap = recap);
    } catch (_) {
      // The recap stays as it was.
    }
  }

  @override
  Widget build(BuildContext context) {
    final settings = _settings!;
    final recap = _recap;
    if (!settings.showProgress ||
        recap == null ||
        recap.weekStart == settings.recapDismissedWeek) {
      return const SizedBox.shrink();
    }
    return _Card(
      recap: recap,
      minimalist: settings.minimalist,
      onDismiss: () => settings.dismissRecap(recap.weekStart),
    );
  }
}

class _Card extends StatelessWidget {
  const _Card({
    required this.recap,
    required this.minimalist,
    required this.onDismiss,
  });

  final golang.Recap recap;
  final bool minimalist;
  final VoidCallback onDismiss;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final last = recap.last;
    final before = recap.before;
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(12),
      side: minimalist
          ? BorderSide(color: colors.outlineVariant)
          : BorderSide.none,
    );
    String change(int n) => switch (n) {
      > 0 => '+${count(n)}',
      < 0 => '−${count(-n)}',
      _ => l.recapSame,
    };
    final retention = last.retention;
    final beforeRetention = before.retention;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: Material(
        key: const Key('weekly-recap'),
        color: minimalist ? Colors.transparent : colors.surfaceContainerHigh,
        shape: shape,
        child: InkWell(
          customBorder: shape,
          onTap: () => showStats(context, collectionScope),
          child: Tooltip(
            message: l.recapTooltip,
            excludeFromSemantics: true,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(12, 4, 4, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.insights_outlined,
                        size: 20,
                        color: colors.primary,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          l.recapTitle,
                          style: theme.textTheme.titleSmall,
                        ),
                      ),
                      IconButton(
                        key: const Key('recap-dismiss'),
                        tooltip: l.recapDismiss,
                        icon: const Icon(Icons.close, size: 20),
                        onPressed: onDismiss,
                      ),
                    ],
                  ),
                  Wrap(
                    spacing: 24,
                    runSpacing: 8,
                    children: [
                      _Metric(
                        key: const Key('recap-days'),
                        value: l.recapStudyDays(last.studyDays),
                        change: change(last.studyDays - before.studyDays),
                        before: l.recapStudyDays(before.studyDays),
                      ),
                      _Metric(
                        key: const Key('recap-answers'),
                        value: l.recapAnswers(last.answers),
                        change: change(last.answers - before.answers),
                        before: l.recapAnswers(before.answers),
                      ),
                      _Metric(
                        key: const Key('recap-minutes'),
                        value: l.recapMinutes(last.minutes),
                        change: change(last.minutes - before.minutes),
                        before: l.recapMinutes(before.minutes),
                      ),
                      if (retention != null)
                        _Metric(
                          key: const Key('recap-retention'),
                          value: l.recapRetention(percent(retention)),
                          change: beforeRetention == null
                              ? null
                              : switch (((retention - beforeRetention) * 100)
                                    .round()) {
                                  0 => l.recapSame,
                                  final n => l.recapPoints(change(n)),
                                },
                          before: beforeRetention == null
                              ? null
                              : percent(beforeRetention),
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// One number of the recap, with its change from the week before.
class _Metric extends StatelessWidget {
  const _Metric({
    super.key,
    required this.value,
    required this.change,
    required this.before,
  });

  final String value;

  /// "+2", "−120" or "same"; null when the week before has no number.
  final String? change;

  /// The week before's number, for the tooltip.
  final String? before;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final change = this.change;
    final before = this.before;
    final text = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(value, style: theme.textTheme.bodyLarge),
        if (change != null)
          Text(
            change,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
              fontFeatures: const [FontFeature.tabularFigures()],
            ),
          ),
      ],
    );
    if (before == null) return text;
    return Tooltip(message: context.l10n.recapWeekBefore(before), child: text);
  }
}
