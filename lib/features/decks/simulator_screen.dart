import 'dart:async';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/features/stats/stats_charts.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Simulates a deck's workload with the FSRS model: the expected reviews
/// per day and memorised cards over a number of days, at a desired
/// retention and optionally a second one to compare. Help me decide
/// suggests the retention with the lowest review time per memorised card.
/// Nothing changes: the screen pops with the retention to apply, after the
/// user confirms, or null. It is opened from the deck options with the
/// options shown there ([options], which may be unsaved).
class SimulatorScreen extends StatefulWidget {
  const SimulatorScreen({super.key, required this.deck, required this.options});

  final golang.DeckNode deck;
  final golang.DeckOptions options;

  @override
  State<SimulatorScreen> createState() => _SimulatorScreenState();
}

class _SimulatorScreenState extends State<SimulatorScreen> {
  final _days = TextEditingController(text: '365');
  late final _newPerDay = TextEditingController(
    text: '${widget.options.newPerDay}',
  );
  late final _reviewsPerDay = TextEditingController(
    text: '${widget.options.reviewsPerDay}',
  );
  late double _retention = _round(
    widget.options.desiredRetention.clamp(0.7, 0.99),
  );
  bool _compare = false;
  late double _compareRetention = _round(
    _retention <= 0.94 ? _retention + 0.05 : _retention - 0.05,
  );

  /// The running simulation or search and its latest progress.
  StreamSubscription<golang.SimulationEvent>? _running;
  golang.SimulationEvent? _progress;
  String? _error;

  /// The last simulation's series and the last suggestion.
  List<golang.SimulationSeries> _series = const [];
  golang.RetentionSuggestion? _suggestion;

  static double _round(double v) => (v * 100).round() / 100;

  @override
  void dispose() {
    _running?.cancel();
    _days.dispose();
    _newPerDay.dispose();
    _reviewsPerDay.dispose();
    super.dispose();
  }

  /// The inputs in the fields. Throws a [FormatException] saying what to
  /// correct first.
  golang.SimulationInput _input() {
    final days = int.tryParse(_days.text.trim());
    final newPerDay = int.tryParse(_newPerDay.text.trim());
    final reviewsPerDay = int.tryParse(_reviewsPerDay.text.trim());
    if (days == null || days < 1 || days > 3650) {
      throw FormatException(context.l10n.simDaysError);
    }
    if (newPerDay == null || reviewsPerDay == null) {
      throw FormatException(context.l10n.simLimitsError);
    }
    return golang.SimulationInput(
      days: days,
      newPerDay: newPerDay,
      reviewsPerDay: reviewsPerDay,
      retention: _retention,
      compareRetention: _compare ? _compareRetention : 0,
    );
  }

  /// Runs [start] with the inputs, showing its progress, and calls [done]
  /// (within setState) with its final event.
  void _run(
    Stream<golang.SimulationEvent> Function(golang.SimulationInput) start,
    void Function(golang.SimulationEvent) done,
  ) {
    final golang.SimulationInput input;
    try {
      input = _input();
    } on FormatException catch (e) {
      setState(() => _error = goText(e.message));
      return;
    }
    setState(() {
      _error = null;
      _progress = null;
    });
    late final StreamSubscription<golang.SimulationEvent> sub;
    sub = start(input).listen(
      (event) {
        if (!mounted || _running != sub) return;
        setState(() {
          if (event.error.isNotEmpty) {
            _error = event.error;
          } else if (event.stage == 'done') {
            done(event);
          } else {
            _progress = event;
          }
        });
      },
      onError: (Object e) {
        if (mounted && _running == sub) {
          setState(() => _error = errorMessage(e));
        }
      },
      onDone: () {
        if (mounted && _running == sub) setState(() => _running = null);
      },
    );
    _running = sub;
  }

  void _simulate() => _run(
    (input) => golang.decks.simulate(widget.deck.id, widget.options, input),
    (event) => _series = event.series,
  );

  void _helpMeDecide() => _run(
    (input) =>
        golang.decks.suggestRetention(widget.deck.id, widget.options, input),
    (event) => _suggestion = event.suggestion,
  );

  void _cancel() {
    _running?.cancel();
    setState(() {
      _running = null;
      _progress = null;
    });
  }

  /// Asks before returning the suggested retention to the deck options.
  Future<void> _apply(double retention) async {
    if (!await confirm(
      context,
      title: context.l10n.useRetentionTitle(percent(retention)),
      message: context.l10n.useRetentionBody,
      action: context.l10n.use,
    )) {
      return;
    }
    if (mounted) Navigator.pop(context, retention);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(l.simulateDeck(widget.deck.name))),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(l.simulatorIntro, style: theme.textTheme.bodySmall),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(child: _number('simDays', _days, l.daysToSimulate)),
              const SizedBox(width: 12),
              Expanded(
                child: _number(
                  'simNewPerDay',
                  _newPerDay,
                  l.newCardsPerDayShort,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _number(
                  'simReviewsPerDay',
                  _reviewsPerDay,
                  l.maxReviewsPerDayShort,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(l.desiredRetentionValue(percent(_retention))),
          Slider(
            key: const Key('simRetention'),
            value: _retention,
            min: 0.70,
            max: 0.99,
            divisions: 29,
            label: percent(_retention),
            onChanged: _running != null
                ? null
                : (v) => setState(() => _retention = _round(v)),
          ),
          SwitchListTile(
            key: const Key('simCompare'),
            contentPadding: EdgeInsets.zero,
            title: Text(l.compareRetention),
            value: _compare,
            onChanged: _running != null
                ? null
                : (v) => setState(() => _compare = v),
          ),
          if (_compare) ...[
            Text(l.comparedRetentionValue(percent(_compareRetention))),
            Slider(
              key: const Key('simCompareRetention'),
              value: _compareRetention,
              min: 0.70,
              max: 0.99,
              divisions: 29,
              label: percent(_compareRetention),
              onChanged: _running != null
                  ? null
                  : (v) => setState(() => _compareRetention = _round(v)),
            ),
          ],
          const SizedBox(height: 8),
          if (_running != null) ...[
            Text(simulationStage(_progress), key: const Key('simProgress')),
            const SizedBox(height: 4),
            LinearProgressIndicator(
              value: _progress == null || _progress!.total == 0
                  ? null
                  : _progress!.done / _progress!.total,
            ),
            Align(
              alignment: Alignment.centerRight,
              child: TextButton(
                key: const Key('simCancel'),
                onPressed: _cancel,
                child: Text(l.cancel),
              ),
            ),
          ] else
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                FilledButton(
                  key: const Key('simulate'),
                  onPressed: _simulate,
                  child: Text(l.simulate),
                ),
                OutlinedButton(
                  key: const Key('helpMeDecide'),
                  onPressed: _helpMeDecide,
                  child: Text(l.helpMeDecide),
                ),
              ],
            ),
          if (_error != null) ...[
            const SizedBox(height: 8),
            Text(
              _error!,
              key: const Key('simError'),
              style: TextStyle(color: theme.colorScheme.error),
            ),
          ],
          if (_suggestion != null) ..._suggestionSection(context, _suggestion!),
          if (_series.isNotEmpty) ..._results(context, _series),
        ],
      ),
    );
  }

  Widget _number(String key, TextEditingController c, String label) =>
      TextField(
        key: Key(key),
        controller: c,
        enabled: _running == null,
        keyboardType: TextInputType.number,
        decoration: InputDecoration(labelText: label),
      );

  /// The suggested retention, its cost curve and Use.
  List<Widget> _suggestionSection(
    BuildContext context,
    golang.RetentionSuggestion s,
  ) {
    final theme = Theme.of(context);
    final colors = StatsColors.of(context);
    final l = context.l10n;
    return [
      const SizedBox(height: 24),
      Text(
        l.suggestedRetention(percent(s.retention)),
        key: const Key('simSuggestion'),
        style: theme.textTheme.titleSmall,
      ),
      const SizedBox(height: 4),
      Text(l.suggestedRetentionHelp, style: theme.textTheme.bodySmall),
      const SizedBox(height: 8),
      Text(l.secondsPerMemorisedCard, style: theme.textTheme.labelMedium),
      const SizedBox(height: 4),
      MultiLineChart(
        key: const Key('simCostChart'),
        lines: [
          [for (final p in s.points) p.secondsPerCard],
        ],
        colors: [colors.line],
        xLabel: (i) =>
            i < s.points.length ? percent(s.points[i].retention) : '',
        height: 160,
      ),
      const SizedBox(height: 8),
      Align(
        alignment: Alignment.centerLeft,
        child: FilledButton.tonal(
          key: const Key('simApply'),
          onPressed: () => _apply(s.retention),
          child: Text(l.useValue(percent(s.retention))),
        ),
      ),
    ];
  }

  /// The charts and totals of a simulation.
  List<Widget> _results(
    BuildContext context,
    List<golang.SimulationSeries> series,
  ) {
    final theme = Theme.of(context);
    final c = StatsColors.of(context);
    final colors = [c.bar, c.line];
    final labels = [for (final s in series) percent(s.retention)];
    final l = context.l10n;
    String day(int i) => l.dayNumber(i + 1);
    return [
      const SizedBox(height: 24),
      if (series.length > 1) ...[
        Legend(labels: labels, colors: colors),
        const SizedBox(height: 8),
      ],
      Text(l.reviewsPerDay, style: theme.textTheme.labelMedium),
      const SizedBox(height: 4),
      MultiLineChart(
        key: const Key('simReviewsChart'),
        lines: [
          for (final s in series) [for (final r in s.reviews) r.toDouble()],
        ],
        colors: colors,
        lineLabels: labels,
        xLabel: day,
        height: 180,
      ),
      const SizedBox(height: 16),
      Text(l.memorisedCards, style: theme.textTheme.labelMedium),
      const SizedBox(height: 4),
      MultiLineChart(
        key: const Key('simMemorizedChart'),
        lines: [for (final s in series) s.memorized],
        colors: colors,
        lineLabels: labels,
        xLabel: day,
        height: 180,
      ),
      const SizedBox(height: 16),
      for (final (i, s) in series.indexed) _totals(context, s, i),
    ];
  }

  Widget _totals(BuildContext context, golang.SimulationSeries s, int i) {
    final days = s.reviews.length;
    final memorized = s.memorized.isEmpty ? 0.0 : s.memorized.last;
    final l = context.l10n;
    final oneDecimal = NumberFormat('0.0', l.localeName);
    return Card(
      key: Key('simTotals$i'),
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l.atRetention(percent(s.retention)),
              style: Theme.of(context).textTheme.titleSmall,
            ),
            const SizedBox(height: 4),
            Text(
              l.simReviewTotals(
                s.totalReviews,
                oneDecimal.format(s.totalReviews / days),
                s.totalNew,
              ),
            ),
            Text(
              l.simTime(
                formatSpent(Duration(seconds: s.totalSeconds.round())),
                formatSpent(Duration(seconds: (s.totalSeconds / days).round())),
              ),
            ),
            Text(l.simMemorisedAtEnd(memorized.round())),
          ],
        ),
      ),
    );
  }
}

/// A retention as a percentage: "90%".
String percent(double retention) =>
    NumberFormat.percentPattern(appL10n.localeName).format(retention);

/// The progress of a simulation, for the screen.
String simulationStage(golang.SimulationEvent? event) {
  if (event == null || event.stage == 'reading') {
    return appL10n.readingDeck;
  }
  if (event.total > 1) {
    return appL10n.simulatingPercent(percent(event.done / event.total));
  }
  return appL10n.simulating;
}
