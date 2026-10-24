import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/features/decks/simulator_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

enum _PresetAction { add, clone, rename, delete, subdecks }

/// Edits a deck's options: which preset it uses and that preset's daily
/// limits and how they apply, learning steps, display order, burying,
/// interval limits, leech handling, desired retention and FSRS parameters,
/// Easy Days, audio, timer and auto advance options, plus the deck's own
/// limits. The
/// preset bar adds, clones, renames and deletes presets and saves the preset
/// to all subdecks. Optimize fits the FSRS parameters to the review history
/// in the background, with progress and Cancel; Simulate opens the workload
/// simulator, which can set the desired retention.
class DeckOptionsScreen extends StatefulWidget {
  const DeckOptionsScreen({super.key, required this.deck});

  final golang.DeckNode deck;

  @override
  State<DeckOptionsScreen> createState() => _DeckOptionsScreenState();
}

class _DeckOptionsScreenState extends State<DeckOptionsScreen> {
  golang.DeckSettings? _settings;
  String? _loadError;

  /// The preset shown and saved; the deck switches to it on save.
  int _presetId = 0;

  /// The options of the shown preset as loaded, to tell unsaved changes.
  late golang.DeckOptions _loaded;

  /// Changes whenever the fields are refilled, so the dropdowns, which only
  /// read their initial value, are rebuilt.
  int _generation = 0;

  final _newPerDay = TextEditingController();
  final _reviewsPerDay = TextEditingController();
  final _deckNewLimit = TextEditingController();
  final _deckReviewLimit = TextEditingController();
  double _retention = 0.9;
  bool _disableAutoplay = false;
  bool _skipQuestionOnReplay = false;
  final _learningSteps = TextEditingController();
  final _relearningSteps = TextEditingController();
  String _interdayOrder = 'mix';
  String _newGatherOrder = 'ascending';
  String _newSortOrder = 'gathered';
  String _newReviewOrder = 'mix';
  String _reviewSortOrder = 'dueRandom';
  String _insertionOrder = 'sequential';
  bool _buryNew = false;
  bool _buryReview = false;
  bool _buryInterday = false;
  final _maxInterval = TextEditingController();
  final _minRelearnInterval = TextEditingController();
  final _leechThreshold = TextEditingController();
  String _leechAction = 'tag';
  bool _newIgnoresReviewLimit = true;
  bool _limitsStartFromTop = false;
  final _maxAnswerSeconds = TextEditingController();
  bool _showTimer = false;
  bool _stopTimerOnAnswer = true;
  final _secondsToShowQuestion = TextEditingController();
  final _secondsToShowAnswer = TextEditingController();
  String _questionAction = 'showAnswer';
  String _answerAction = 'buryCard';
  final _weights = TextEditingController();
  final _optimizeSearch = TextEditingController();
  String _ignoreBefore = '';
  double _historicalRetention = 0.9;
  bool _healthCheck = false;
  bool _rescheduleOnChange = false;

  /// The Easy Days level of each weekday, Monday first (see [easyDayLevels]).
  List<int> _easyDays = List.filled(7, 0);
  String? _error;
  bool _busy = false;

  /// The running Optimize or Optimize all, its latest progress, and for
  /// Optimize all which preset it is on.
  StreamSubscription<golang.OptimizeEvent>? _optimizing;
  golang.OptimizeEvent? _optimizeProgress;
  String _optimizePreset = '';

  /// What the last Optimize found.
  String? _optimizeMessage;

  bool get _locked => _busy || _optimizing != null;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    for (final c in [
      _newPerDay,
      _reviewsPerDay,
      _deckNewLimit,
      _deckReviewLimit,
      _learningSteps,
      _relearningSteps,
      _maxInterval,
      _minRelearnInterval,
      _leechThreshold,
      _maxAnswerSeconds,
      _secondsToShowQuestion,
      _secondsToShowAnswer,
      _weights,
      _optimizeSearch,
    ]) {
      c.dispose();
    }
    _optimizing?.cancel();
    super.dispose();
  }

  /// Loads the deck's settings and shows its preset, or [show] when given
  /// and it still exists. With [keepFields] only the preset list is
  /// refreshed.
  Future<void> _load({int? show, bool keepFields = false}) async {
    try {
      final settings = await golang.decks.settings(widget.deck.id);
      if (!mounted) return;
      setState(() {
        _settings = settings;
        _loadError = null;
        if (keepFields && _preset(_presetId) != null) return;
        final id = show != null && _presetIn(settings, show) != null
            ? show
            : settings.presetID;
        _deckNewLimit.text = _limitText(settings.limits.newPerDay);
        _deckReviewLimit.text = _limitText(settings.limits.reviewsPerDay);
        _show(_presetIn(settings, id)!);
      });
    } catch (e) {
      if (mounted) setState(() => _loadError = errorMessage(e));
    }
  }

  static String _limitText(int? limit) => limit == null ? '' : '$limit';

  static golang.Preset? _presetIn(golang.DeckSettings s, int id) {
    for (final p in s.presets) {
      if (p.id == id) return p;
    }
    return null;
  }

  golang.Preset? _preset(int id) =>
      _settings == null ? null : _presetIn(_settings!, id);

  /// Fills the fields with [preset]'s options.
  void _show(golang.Preset preset) {
    final o = preset.options;
    _presetId = preset.id;
    _loaded = o;
    _generation++;
    _error = null;
    _newPerDay.text = '${o.newPerDay}';
    _reviewsPerDay.text = '${o.reviewsPerDay}';
    _retention = o.desiredRetention.clamp(0.7, 0.99);
    _disableAutoplay = o.disableAutoplay;
    _skipQuestionOnReplay = o.skipQuestionOnReplay;
    _learningSteps.text = formatSteps(o.learningSteps);
    _relearningSteps.text = formatSteps(o.relearningSteps);
    _interdayOrder = o.interdayLearningOrder;
    _newGatherOrder = o.newGatherOrder;
    _newSortOrder = o.newSortOrder;
    _newReviewOrder = o.newReviewOrder;
    _reviewSortOrder = o.reviewSortOrder;
    _insertionOrder = o.insertionOrder;
    _buryNew = o.buryNew;
    _buryReview = o.buryReview;
    _buryInterday = o.buryInterdayLearning;
    _maxInterval.text = '${o.maxInterval}';
    _minRelearnInterval.text = '${o.minRelearnInterval}';
    _leechThreshold.text = '${o.leechThreshold}';
    _leechAction = o.leechAction;
    _newIgnoresReviewLimit = o.newIgnoresReviewLimit;
    _limitsStartFromTop = o.limitsStartFromTop;
    _maxAnswerSeconds.text = '${o.maxAnswerSeconds}';
    _showTimer = o.showTimer;
    _stopTimerOnAnswer = o.stopTimerOnAnswer;
    _secondsToShowQuestion.text = '${o.secondsToShowQuestion}';
    _secondsToShowAnswer.text = '${o.secondsToShowAnswer}';
    _questionAction = o.questionAction;
    _answerAction = o.answerAction;
    _weights.text = formatWeights(o.weights);
    _optimizeSearch.text = o.optimizeSearch;
    _ignoreBefore = o.ignoreBefore;
    _historicalRetention = o.historicalRetention.clamp(0.5, 0.99);
    _healthCheck = o.healthCheck;
    _rescheduleOnChange = o.rescheduleOnChange;
    _easyDays = [
      for (var i = 0; i < 7; i++) i < o.easyDays.length ? o.easyDays[i] : 0,
    ];
    _optimizeMessage = null;
  }

  /// The problem with the FSRS parameters, or null when they are valid.
  String? get _weightsError {
    try {
      parseWeights(_weights.text);
      return null;
    } on FormatException catch (e) {
      return e.message;
    }
  }

  /// The problem with the steps in [c], or null when they are valid.
  static String? _stepsError(TextEditingController c) {
    try {
      parseSteps(c.text);
      return null;
    } on FormatException catch (e) {
      return e.message;
    }
  }

  /// The number of days in [c], or null when it is not a whole number from
  /// 1 to 36,500.
  static int? _days(TextEditingController c) {
    final days = int.tryParse(c.text.trim());
    return days != null && days >= 1 && days <= maxIntervalDays ? days : null;
  }

  String? get _maxIntervalError =>
      _days(_maxInterval) == null ? context.l10n.enterDaysError : null;

  String? get _minRelearnIntervalError {
    final min = _days(_minRelearnInterval);
    if (min == null) return context.l10n.enterDaysError;
    final max = _days(_maxInterval);
    if (max != null && min > max) {
      return context.l10n.notAboveMaxInterval;
    }
    return null;
  }

  /// The seconds in [_maxAnswerSeconds], or null when it is not a whole
  /// number from 1 to 9999.
  int? get _maxAnswer {
    final seconds = int.tryParse(_maxAnswerSeconds.text.trim());
    return seconds != null && seconds >= 1 && seconds <= maxAnswerSeconds
        ? seconds
        : null;
  }

  String? get _maxAnswerError => _maxAnswer == null
      ? context.l10n.enterSecondsError(1, maxAnswerSeconds)
      : null;

  /// The seconds in [c], or null when it is not a whole number from 0 to
  /// 9999.
  static int? _autoAdvanceSeconds(TextEditingController c) {
    final seconds = int.tryParse(c.text.trim());
    return seconds != null && seconds >= 0 && seconds <= maxAutoAdvanceSeconds
        ? seconds
        : null;
  }

  static String? _autoAdvanceError(TextEditingController c) =>
      _autoAdvanceSeconds(c) == null
      ? appL10n.enterSecondsError(0, maxAutoAdvanceSeconds)
      : null;

  /// A deck limit field's value: null when empty (use the preset's).
  static int? _limit(TextEditingController c) {
    final text = c.text.trim();
    return text.isEmpty ? null : int.tryParse(text);
  }

  static bool _limitValid(TextEditingController c) =>
      c.text.trim().isEmpty || int.tryParse(c.text.trim()) != null;

  /// The options in the fields. Throws a [FormatException] saying what to
  /// correct first.
  golang.DeckOptions _collect() {
    final newPerDay = int.tryParse(_newPerDay.text.trim());
    final reviewsPerDay = int.tryParse(_reviewsPerDay.text.trim());
    final l = context.l10n;
    if (newPerDay == null || reviewsPerDay == null) {
      throw FormatException(l.simLimitsError);
    }
    final leechThreshold = int.tryParse(_leechThreshold.text.trim());
    if (leechThreshold == null) {
      throw FormatException(l.leechThresholdError);
    }
    if (_stepsError(_learningSteps) != null ||
        _stepsError(_relearningSteps) != null) {
      throw FormatException(l.correctSteps);
    }
    if (_maxIntervalError != null || _minRelearnIntervalError != null) {
      throw FormatException(l.correctIntervals);
    }
    if (_maxAnswerError != null) {
      throw FormatException(l.correctMaxAnswer);
    }
    if (_autoAdvanceError(_secondsToShowQuestion) != null ||
        _autoAdvanceError(_secondsToShowAnswer) != null) {
      throw FormatException(l.correctAutoAdvance);
    }
    if (_weightsError != null) {
      throw FormatException(l.correctFsrs);
    }
    return golang.DeckOptions(
      newPerDay: newPerDay,
      reviewsPerDay: reviewsPerDay,
      desiredRetention: double.parse(_retention.toStringAsFixed(2)),
      disableAutoplay: _disableAutoplay,
      skipQuestionOnReplay: _skipQuestionOnReplay,
      learningSteps: parseSteps(_learningSteps.text),
      relearningSteps: parseSteps(_relearningSteps.text),
      interdayLearningOrder: _interdayOrder,
      buryNew: _buryNew,
      buryReview: _buryReview,
      buryInterdayLearning: _buryInterday,
      leechThreshold: leechThreshold,
      leechAction: _leechAction,
      newIgnoresReviewLimit: _newIgnoresReviewLimit,
      limitsStartFromTop: _limitsStartFromTop,
      maxInterval: _days(_maxInterval)!,
      minRelearnInterval: _days(_minRelearnInterval)!,
      newGatherOrder: _newGatherOrder,
      newSortOrder: _newSortOrder,
      newReviewOrder: _newReviewOrder,
      reviewSortOrder: _reviewSortOrder,
      insertionOrder: _insertionOrder,
      maxAnswerSeconds: _maxAnswer!,
      showTimer: _showTimer,
      stopTimerOnAnswer: _stopTimerOnAnswer,
      secondsToShowQuestion: _autoAdvanceSeconds(_secondsToShowQuestion)!,
      secondsToShowAnswer: _autoAdvanceSeconds(_secondsToShowAnswer)!,
      questionAction: _questionAction,
      answerAction: _answerAction,
      weights: parseWeights(_weights.text),
      optimizeSearch: _optimizeSearch.text.trim(),
      ignoreBefore: _ignoreBefore,
      historicalRetention: double.parse(
        _historicalRetention.toStringAsFixed(2),
      ),
      healthCheck: _healthCheck,
      rescheduleOnChange: _rescheduleOnChange,
      easyDays: List.of(_easyDays),
    );
  }

  /// Whether the fields differ from the shown preset's saved options.
  bool get _changed {
    try {
      return jsonEncode(_collect().toJson()) != jsonEncode(_loaded.toJson());
    } on FormatException {
      return true;
    }
  }

  /// Saves the fields: the deck uses the shown preset, which gets the
  /// options, and the deck's own limits. Returns whether it succeeded;
  /// otherwise the reason is shown.
  Future<bool> _save() async {
    final golang.DeckOptions options;
    try {
      options = _collect();
    } on FormatException catch (e) {
      setState(() => _error = e.message);
      return false;
    }
    if (!_limitValid(_deckNewLimit) || !_limitValid(_deckReviewLimit)) {
      setState(() => _error = context.l10n.deckLimitsError);
      return false;
    }
    try {
      final reschedule = await golang.decks.rescheduleCount(
        widget.deck.id,
        _presetId,
        options,
      );
      if (reschedule > 0 &&
          mounted &&
          !await confirm(
            context,
            title: context.l10n.rescheduleTitle,
            message: context.l10n.rescheduleBody(reschedule),
            action: context.l10n.rescheduleAction,
          )) {
        return false;
      }
      await golang.decks.saveSettings(
        widget.deck.id,
        _presetId,
        options,
        golang.DeckLimits(
          newPerDay: _limit(_deckNewLimit),
          reviewsPerDay: _limit(_deckReviewLimit),
        ),
      );
      return true;
    } catch (e) {
      setState(() => _error = errorMessage(e));
      return false;
    }
  }

  /// Fits the FSRS parameters to the review history the fields select.
  /// Better parameters go into the parameters field, to be saved.
  void _optimize() {
    final golang.DeckOptions options;
    try {
      options = _collect();
    } on FormatException catch (e) {
      setState(() => _error = e.message);
      return;
    }
    _run(golang.decks.optimize(widget.deck.id, options), (event) {
      final r = event.result!;
      if (r.use) _weights.text = formatWeights(r.weights);
      _optimizeMessage = optimizeMessage(r);
    });
  }

  /// Opens the simulator with the options in the fields; a retention it
  /// returns goes into the desired retention, to be saved.
  Future<void> _simulate() async {
    final golang.DeckOptions options;
    try {
      options = _collect();
    } on FormatException catch (e) {
      setState(() => _error = e.message);
      return;
    }
    final retention = await Navigator.push<double>(
      context,
      MaterialPageRoute(
        builder: (_) => SimulatorScreen(deck: widget.deck, options: options),
      ),
    );
    if (retention != null && mounted) {
      setState(() => _retention = retention.clamp(0.7, 0.99));
    }
  }

  /// Optimizes every preset with its saved options, after saving the
  /// fields and a confirmation, and shows the result.
  Future<void> _optimizeAll() async {
    if (_changed && !await _save()) return;
    final reschedule = await golang.decks.rescheduleCountAll();
    if (!mounted) return;
    final l = context.l10n;
    if (!await confirm(
      context,
      title: l.optimizeAllTitle,
      message:
          l.optimizeAllBody +
          (reschedule == 0 ? '' : l.optimizeAllReschedule(reschedule)),
      action: l.optimize,
    )) {
      return;
    }
    _run(golang.decks.optimizeAll(), (event) {
      final applied = event.presets.where((p) => p.applied).length;
      final few = event.presets.where((p) => p.result.notEnoughReviews).length;
      final message =
          l.optimizeAllResult(applied) +
          (few == 0 ? '' : l.optimizeAllFew(few));
      // Reloading shows the new parameters, then the result.
      _load(show: _presetId).then((_) {
        if (mounted) setState(() => _optimizeMessage = message);
      });
    });
  }

  /// Runs an optimization [events] stream, showing its progress, and calls
  /// [done] (within setState) with its final event.
  void _run(
    Stream<golang.OptimizeEvent> events,
    void Function(golang.OptimizeEvent) done,
  ) {
    setState(() {
      _error = null;
      _optimizeMessage = null;
      _optimizePreset = '';
      _optimizeProgress = null;
    });
    late final StreamSubscription<golang.OptimizeEvent> sub;
    sub = events.listen(
      (event) {
        if (!mounted || _optimizing != sub) return;
        setState(() {
          if (event.error.isNotEmpty) {
            _optimizeMessage = event.error;
          } else if (event.stage == 'done') {
            done(event);
          } else if (event.stage == 'preset') {
            _optimizePreset = event.done < event.total
                ? context.l10n.optimizePreset(event.done + 1, event.total)
                : '';
          } else {
            _optimizeProgress = event;
          }
        });
      },
      onError: (Object e) {
        if (mounted && _optimizing == sub) {
          setState(() => _optimizeMessage = errorMessage(e));
        }
      },
      onDone: () {
        if (mounted && _optimizing == sub) setState(() => _optimizing = null);
      },
    );
    _optimizing = sub;
  }

  void _cancelOptimize() {
    _optimizing?.cancel();
    setState(() {
      _optimizing = null;
      _optimizeMessage = context.l10n.optimizeCancelled;
    });
  }

  Future<void> _pickIgnoreBefore() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: DateTime.tryParse(_ignoreBefore) ?? now,
      firstDate: DateTime(2006),
      lastDate: now,
      helpText: context.l10n.ignoreReviewedBefore,
    );
    if (picked != null) {
      setState(() => _ignoreBefore = formatDate(picked));
    }
  }

  Future<void> _saveAndClose() async {
    if (await _save() && mounted) Navigator.pop(context, true);
  }

  /// Shows another preset, after asking to discard unsaved changes.
  Future<void> _choose(int id) async {
    if (id == _presetId) return;
    final name = _preset(_presetId)?.name ?? '';
    if (_changed &&
        !await confirm(
          context,
          title: context.l10n.discardChangesTitle,
          message: context.l10n.presetNotSaved(name),
          action: context.l10n.discard,
        )) {
      setState(() => _generation++); // put the picker back
      return;
    }
    setState(() => _show(_preset(id)!));
  }

  /// Runs a preset action, then reloads the screen showing [show] (the
  /// deck's preset when null). With [saveFirst] the fields are saved
  /// first, so the action sees them; it does not run when that fails.
  Future<void> _act(
    Future<int?> Function() action, {
    bool saveFirst = false,
    bool keepFields = false,
  }) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      if (saveFirst && !await _save()) return;
      final show = await action();
      await _load(show: show, keepFields: keepFields);
    } catch (e) {
      showError(e);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _presetAction(_PresetAction action) async {
    final preset = _preset(_presetId)!;
    final l = context.l10n;
    switch (action) {
      case _PresetAction.add:
        final name = await askText(
          context,
          title: l.addPreset,
          label: l.name,
          helper: l.addPresetHelper,
          action: l.add,
        );
        if (name == null) return;
        await _act(
          () async => await golang.decks.addPreset(widget.deck.id, name),
          saveFirst: _changed,
        );
      case _PresetAction.clone:
        final name = await askText(
          context,
          title: l.clonePreset,
          initial: l.presetCopyName(preset.name),
          label: l.name,
          helper: l.clonePresetHelper,
          action: l.clone,
        );
        if (name == null) return;
        await _act(
          () async => await golang.decks.clonePreset(widget.deck.id, name),
          saveFirst: true,
        );
      case _PresetAction.rename:
        final name = await askText(
          context,
          title: l.renamePreset,
          initial: preset.name,
          label: l.name,
          action: l.rename,
        );
        if (name == null) return;
        await _act(() async {
          await golang.decks.renamePreset(preset.id, name);
          return preset.id;
        }, keepFields: true);
      case _PresetAction.delete:
        final uses = preset.decks;
        if (!await confirm(
          context,
          title: l.deletePresetTitle,
          message: uses == 0
              ? l.presetUnused(preset.name)
              : l.presetUsedBy(
                  preset.name,
                  uses,
                  _preset(_settings!.defaultPresetID)?.name ??
                      l.builtinPresetName,
                ),
          action: l.delete,
          destructive: true,
        )) {
          return;
        }
        await _act(() async {
          await golang.decks.deletePreset(preset.id);
          return null;
        });
      case _PresetAction.subdecks:
        await _act(() async {
          await golang.decks.savePresetToSubdecks(widget.deck.id);
          showInfo(l.allSubdecksUsePreset(preset.name));
          return preset.id;
        }, saveFirst: true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final settings = _settings;
    return Scaffold(
      appBar: AppBar(
        title: Text(context.l10n.optionsFor(widget.deck.name)),
        actions: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: FilledButton(
              onPressed: settings == null || _locked ? null : _saveAndClose,
              child: Text(context.l10n.save),
            ),
          ),
        ],
      ),
      body: settings == null
          ? Center(
              child: _loadError == null
                  ? const CircularProgressIndicator()
                  : Padding(
                      padding: const EdgeInsets.all(16),
                      child: Text(_loadError!),
                    ),
            )
          : Align(
              alignment: Alignment.topCenter,
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 640),
                child: SingleChildScrollView(
                  key: ValueKey(_generation),
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _presetBar(settings),
                      const SizedBox(height: 16),
                      ..._fields(context),
                    ],
                  ),
                ),
              ),
            ),
    );
  }

  /// The preset picker, its actions and how many decks use it.
  Widget _presetBar(golang.DeckSettings settings) {
    final preset = _preset(_presetId)!;
    // The decks using it once saved: this deck counts when it switches.
    final uses = preset.decks + (_presetId == settings.presetID ? 0 : 1);
    final l = context.l10n;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<int>(
                key: const Key('preset'),
                initialValue: _presetId,
                isExpanded: true,
                decoration: InputDecoration(labelText: l.preset),
                items: [
                  for (final p in settings.presets)
                    DropdownMenuItem(value: p.id, child: Text(p.name)),
                ],
                onChanged: _locked ? null : (v) => _choose(v!),
              ),
            ),
            PopupMenuButton<_PresetAction>(
              popUpAnimationStyle: popupAnimation(context),
              tooltip: l.presetActions,
              enabled: !_locked,
              onSelected: _presetAction,
              itemBuilder: (_) => [
                PopupMenuItem(
                  value: _PresetAction.add,
                  child: Text(l.addPreset),
                ),
                PopupMenuItem(
                  value: _PresetAction.clone,
                  child: Text(l.clonePreset),
                ),
                PopupMenuItem(
                  value: _PresetAction.rename,
                  child: Text(l.renamePreset),
                ),
                PopupMenuItem(
                  value: _PresetAction.delete,
                  enabled: preset.id != settings.defaultPresetID,
                  child: Text(l.deletePreset),
                ),
                PopupMenuItem(
                  value: _PresetAction.subdecks,
                  enabled: widget.deck.hasChildren,
                  child: Text(l.saveToAllSubdecks),
                ),
              ],
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          uses > 1 ? l.usedByDecksShared(uses) : l.usedByDecks(uses),
          key: const Key('presetUsage'),
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }

  List<Widget> _fields(BuildContext context) {
    final l = context.l10n;
    return [
      Text(l.dailyLimits, style: Theme.of(context).textTheme.titleSmall),
      TextField(
        key: const Key('newPerDay'),
        controller: _newPerDay,
        keyboardType: TextInputType.number,
        decoration: InputDecoration(labelText: l.newCardsPerDay),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('reviewsPerDay'),
        controller: _reviewsPerDay,
        keyboardType: TextInputType.number,
        decoration: InputDecoration(labelText: l.maxReviewsPerDay),
      ),
      const SizedBox(height: 12),
      Text(l.thisDeckOnly, style: Theme.of(context).textTheme.labelLarge),
      Text(l.thisDeckOnlyHelp, style: Theme.of(context).textTheme.bodySmall),
      Row(
        children: [
          Expanded(
            child: TextField(
              key: const Key('deckNewLimit'),
              controller: _deckNewLimit,
              keyboardType: TextInputType.number,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                labelText: l.newCardsPerDay,
                errorText: _limitValid(_deckNewLimit)
                    ? null
                    : l.enterWholeNumberShort,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              key: const Key('deckReviewLimit'),
              controller: _deckReviewLimit,
              keyboardType: TextInputType.number,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                labelText: l.maxReviewsPerDay,
                errorText: _limitValid(_deckReviewLimit)
                    ? null
                    : l.enterWholeNumberShort,
              ),
            ),
          ),
        ],
      ),
      SwitchListTile(
        key: const Key('newIgnoresReviewLimit'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.newIgnoresReviewLimit),
        subtitle: Text(l.newIgnoresReviewLimitHelp),
        value: _newIgnoresReviewLimit,
        onChanged: (v) => setState(() => _newIgnoresReviewLimit = v),
      ),
      SwitchListTile(
        key: const Key('limitsStartFromTop'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.limitsStartFromTop),
        subtitle: Text(l.limitsStartFromTopHelp),
        value: _limitsStartFromTop,
        onChanged: (v) => setState(() => _limitsStartFromTop = v),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('learningSteps'),
        controller: _learningSteps,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.learningSteps,
          helperText: l.learningStepsHelp,
          helperMaxLines: 3,
          errorText: _stepsError(_learningSteps),
          errorMaxLines: 3,
        ),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('relearningSteps'),
        controller: _relearningSteps,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.relearningSteps,
          helperText: l.relearningStepsHelp,
          helperMaxLines: 3,
          errorText: _stepsError(_relearningSteps),
          errorMaxLines: 3,
        ),
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'insertionOrder',
        label: l.insertionOrder,
        helper: l.insertionOrderHelp,
        value: _insertionOrder,
        options: {'sequential': l.insertSequential, 'random': l.orderRandom},
        onChanged: (v) => _insertionOrder = v,
      ),
      const SizedBox(height: 16),
      Text(l.displayOrder, style: Theme.of(context).textTheme.titleSmall),
      const SizedBox(height: 4),
      _orderDropdown(
        key: 'newGatherOrder',
        label: l.newGatherOrder,
        helper: l.newGatherOrderHelp,
        value: _newGatherOrder,
        options: {
          'deck': l.gatherDeck,
          'deckRandomNotes': l.gatherDeckRandomNotes,
          'ascending': l.gatherAscending,
          'descending': l.gatherDescending,
          'randomNotes': l.gatherRandomNotes,
          'randomCards': l.gatherRandomCards,
        },
        onChanged: (v) => _newGatherOrder = v,
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'newSortOrder',
        label: l.newSortOrder,
        helper: l.newSortOrderHelp,
        value: _newSortOrder,
        options: {
          'gathered': l.sortGathered,
          'typeThenGathered': l.sortTypeThenGathered,
          'typeThenRandom': l.sortTypeThenRandom,
          'randomNoteThenType': l.sortRandomNoteThenType,
          'random': l.orderRandom,
        },
        onChanged: (v) => _newSortOrder = v,
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'newReviewOrder',
        label: l.newReviewOrder,
        value: _newReviewOrder,
        options: {
          'mix': l.mixWithReviews,
          'before': l.showBeforeReviews,
          'after': l.showAfterReviews,
        },
        onChanged: (v) => _newReviewOrder = v,
      ),
      const SizedBox(height: 12),
      DropdownButtonFormField<String>(
        key: const Key('interdayOrder'),
        initialValue: _interdayOrder,
        isExpanded: true,
        decoration: InputDecoration(
          labelText: l.interdayOrder,
          helperText: l.interdayOrderHelp,
          helperMaxLines: 2,
        ),
        items: [
          DropdownMenuItem(value: 'mix', child: Text(l.mixWithReviews)),
          DropdownMenuItem(value: 'before', child: Text(l.showBeforeReviews)),
          DropdownMenuItem(value: 'after', child: Text(l.showAfterReviews)),
        ],
        onChanged: (v) => setState(() => _interdayOrder = v!),
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'reviewSortOrder',
        label: l.reviewSortOrder,
        helper: l.reviewSortOrderHelp,
        value: _reviewSortOrder,
        options: {
          'dueRandom': l.reviewDueRandom,
          'dueDeck': l.reviewDueDeck,
          'deckDue': l.reviewDeckDue,
          'intervalsAscending': l.reviewIntervalsAscending,
          'intervalsDescending': l.reviewIntervalsDescending,
          'difficultyAscending': l.reviewDifficultyAscending,
          'difficultyDescending': l.reviewDifficultyDescending,
          'retrievabilityAscending': l.reviewRetrievabilityAscending,
          'relativeOverdueness': l.reviewRelativeOverdueness,
        },
        onChanged: (v) => _reviewSortOrder = v,
      ),
      const SizedBox(height: 16),
      Text(l.burying, style: Theme.of(context).textTheme.titleSmall),
      Text(l.buryingHelp, style: Theme.of(context).textTheme.bodySmall),
      SwitchListTile(
        key: const Key('buryNew'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.buryNew),
        value: _buryNew,
        onChanged: (v) => setState(() => _buryNew = v),
      ),
      SwitchListTile(
        key: const Key('buryReview'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.buryReview),
        value: _buryReview,
        onChanged: (v) => setState(() => _buryReview = v),
      ),
      SwitchListTile(
        key: const Key('buryInterdayLearning'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.buryInterday),
        value: _buryInterday,
        onChanged: (v) => setState(() => _buryInterday = v),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('maxInterval'),
        controller: _maxInterval,
        keyboardType: TextInputType.number,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.maxIntervalDays,
          helperText: l.maxIntervalHelp,
          helperMaxLines: 2,
          errorText: _maxIntervalError,
          errorMaxLines: 2,
        ),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('minRelearnInterval'),
        controller: _minRelearnInterval,
        keyboardType: TextInputType.number,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.minIntervalDays,
          helperText: l.minIntervalHelp,
          helperMaxLines: 2,
          errorText: _minRelearnIntervalError,
          errorMaxLines: 2,
        ),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('leechThreshold'),
        controller: _leechThreshold,
        keyboardType: TextInputType.number,
        decoration: InputDecoration(
          labelText: l.leechThreshold,
          helperText: l.leechThresholdHelp,
          helperMaxLines: 3,
        ),
      ),
      const SizedBox(height: 12),
      DropdownButtonFormField<String>(
        key: const Key('leechAction'),
        initialValue: _leechAction,
        isExpanded: true,
        decoration: InputDecoration(labelText: l.leechAction),
        items: [
          DropdownMenuItem(value: 'tag', child: Text(l.leechTagOnly)),
          DropdownMenuItem(value: 'suspend', child: Text(l.leechSuspend)),
        ],
        onChanged: (v) => setState(() => _leechAction = v!),
      ),
      const SizedBox(height: 8),
      Text(l.desiredRetentionValue(percent(_retention))),
      Slider(
        key: const Key('retention'),
        value: _retention,
        min: 0.70,
        max: 0.99,
        divisions: 29,
        label: percent(_retention),
        onChanged: (v) => setState(() => _retention = v),
      ),
      Text(l.retentionHelp, style: Theme.of(context).textTheme.bodySmall),
      ..._fsrsFields(context),
      ..._easyDaysFields(context),
      const SizedBox(height: 12),
      SwitchListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(l.disableAutoplay),
        value: _disableAutoplay,
        onChanged: (v) => setState(() => _disableAutoplay = v),
      ),
      SwitchListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(l.skipQuestionOnReplay),
        value: _skipQuestionOnReplay,
        onChanged: (v) => setState(() => _skipQuestionOnReplay = v),
      ),
      const SizedBox(height: 16),
      Text(l.timer, style: Theme.of(context).textTheme.titleSmall),
      const SizedBox(height: 4),
      TextField(
        key: const Key('maxAnswerSeconds'),
        controller: _maxAnswerSeconds,
        keyboardType: TextInputType.number,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.maxAnswerSeconds,
          helperText: l.maxAnswerSecondsHelp,
          helperMaxLines: 3,
          errorText: _maxAnswerError,
          errorMaxLines: 2,
        ),
      ),
      SwitchListTile(
        key: const Key('showTimer'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.showTimer),
        subtitle: Text(l.showTimerHelp),
        value: _showTimer,
        onChanged: (v) => setState(() => _showTimer = v),
      ),
      SwitchListTile(
        key: const Key('stopTimerOnAnswer'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.stopTimerOnAnswer),
        subtitle: Text(l.stopTimerOnAnswerHelp),
        value: _stopTimerOnAnswer,
        onChanged: _showTimer
            ? (v) => setState(() => _stopTimerOnAnswer = v)
            : null,
      ),
      const SizedBox(height: 16),
      Text(l.autoAdvance, style: Theme.of(context).textTheme.titleSmall),
      Text(l.autoAdvanceHelp, style: Theme.of(context).textTheme.bodySmall),
      const SizedBox(height: 4),
      Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: TextField(
              key: const Key('secondsToShowQuestion'),
              controller: _secondsToShowQuestion,
              keyboardType: TextInputType.number,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                labelText: l.secondsToShowQuestion,
                errorText: _autoAdvanceError(_secondsToShowQuestion),
                errorMaxLines: 3,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              key: const Key('secondsToShowAnswer'),
              controller: _secondsToShowAnswer,
              keyboardType: TextInputType.number,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                labelText: l.secondsToShowAnswer,
                errorText: _autoAdvanceError(_secondsToShowAnswer),
                errorMaxLines: 3,
              ),
            ),
          ),
        ],
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'questionAction',
        label: l.questionAction,
        value: _questionAction,
        options: {'showAnswer': l.showAnswer, 'showReminder': l.showReminder},
        onChanged: (v) => _questionAction = v,
      ),
      const SizedBox(height: 12),
      _orderDropdown(
        key: 'answerAction',
        label: l.answerAction,
        value: _answerAction,
        options: {
          'buryCard': l.buryCard,
          'answerAgain': l.answerAgain,
          'answerHard': l.answerHard,
          'answerGood': l.answerGood,
          'showReminder': l.showReminder,
        },
        onChanged: (v) => _answerAction = v,
      ),
      if (_error != null) ...[
        const SizedBox(height: 12),
        Text(
          _error!,
          key: const Key('optionsError'),
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      ],
    ];
  }

  /// The FSRS parameters, what optimizing learns from, Optimize and
  /// rescheduling on change.
  List<Widget> _fsrsFields(BuildContext context) {
    final theme = Theme.of(context);
    final progress = _optimizeProgress;
    final l = context.l10n;
    return [
      const SizedBox(height: 16),
      Text(l.fsrs, style: theme.textTheme.titleSmall),
      const SizedBox(height: 4),
      TextField(
        key: const Key('fsrsWeights'),
        controller: _weights,
        minLines: 1,
        maxLines: 4,
        onChanged: (_) => setState(() {}),
        decoration: InputDecoration(
          labelText: l.fsrsParameters,
          helperText: l.fsrsParametersHelp,
          helperMaxLines: 2,
          errorText: _weightsError,
          errorMaxLines: 2,
        ),
      ),
      const SizedBox(height: 12),
      TextField(
        key: const Key('optimizeSearch'),
        controller: _optimizeSearch,
        decoration: InputDecoration(
          labelText: l.optimizeSearch,
          helperText: l.optimizeSearchHelp,
          helperMaxLines: 2,
        ),
      ),
      ListTile(
        key: const Key('ignoreBefore'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.ignoreReviewedBefore),
        subtitle: Text(_ignoreBefore.isEmpty ? l.notSet : _ignoreBefore),
        onTap: _locked ? null : _pickIgnoreBefore,
        trailing: _ignoreBefore.isEmpty
            ? const Icon(Icons.calendar_today_outlined)
            : IconButton(
                tooltip: l.clearDate,
                icon: const Icon(Icons.clear),
                onPressed: () => setState(() => _ignoreBefore = ''),
              ),
      ),
      Text(l.historicalRetentionValue(percent(_historicalRetention))),
      Slider(
        key: const Key('historicalRetention'),
        value: _historicalRetention,
        min: 0.5,
        max: 0.99,
        divisions: 49,
        label: percent(_historicalRetention),
        onChanged: (v) => setState(() => _historicalRetention = v),
      ),
      Text(l.historicalRetentionHelp, style: theme.textTheme.bodySmall),
      SwitchListTile(
        key: const Key('healthCheck'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.healthCheck),
        subtitle: Text(l.healthCheckHelp),
        value: _healthCheck,
        onChanged: (v) => setState(() => _healthCheck = v),
      ),
      SwitchListTile(
        key: const Key('rescheduleOnChange'),
        contentPadding: EdgeInsets.zero,
        title: Text(l.rescheduleOnChange),
        subtitle: Text(l.rescheduleOnChangeHelp),
        value: _rescheduleOnChange,
        onChanged: (v) => setState(() => _rescheduleOnChange = v),
      ),
      const SizedBox(height: 8),
      if (_optimizing != null) ...[
        Text(
          '$_optimizePreset${optimizeStage(progress)}',
          key: const Key('optimizeProgress'),
        ),
        const SizedBox(height: 4),
        LinearProgressIndicator(
          value: progress == null || progress.total == 0
              ? null
              : progress.done / progress.total,
        ),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton(
            key: const Key('cancelOptimize'),
            onPressed: _cancelOptimize,
            child: Text(l.cancel),
          ),
        ),
      ] else
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            OutlinedButton(
              key: const Key('optimize'),
              onPressed: _busy ? null : _optimize,
              child: Text(l.optimize),
            ),
            OutlinedButton(
              key: const Key('optimizeAll'),
              onPressed: _busy ? null : _optimizeAll,
              child: Text(l.optimizeAllPresets),
            ),
            OutlinedButton(
              key: const Key('simulateWorkload'),
              onPressed: _busy ? null : _simulate,
              child: Text(l.simulate),
            ),
          ],
        ),
      if (_optimizeMessage != null) ...[
        const SizedBox(height: 8),
        Text(_optimizeMessage!, key: const Key('optimizeMessage')),
      ],
    ];
  }

  /// A Normal, Reduced or Minimum choice for each weekday.
  List<Widget> _easyDaysFields(BuildContext context) {
    final theme = Theme.of(context);
    final l = context.l10n;
    final weekdays = weekdayAbbreviations(l);
    return [
      const SizedBox(height: 16),
      Text(l.easyDays, style: theme.textTheme.titleSmall),
      Text(l.easyDaysHelp, style: theme.textTheme.bodySmall),
      const SizedBox(height: 8),
      for (var i = 0; i < 7; i++)
        Padding(
          padding: const EdgeInsets.only(bottom: 6),
          child: Row(
            children: [
              SizedBox(width: 44, child: Text(weekdays[i])),
              Expanded(
                child: SegmentedButton<int>(
                  key: Key('easyDay$i'),
                  expandedInsets: EdgeInsets.zero,
                  showSelectedIcon: false,
                  style: const ButtonStyle(
                    visualDensity: VisualDensity.compact,
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                  segments: [
                    for (final (level, label) in easyDayLevels(l).indexed)
                      ButtonSegment(
                        value: level,
                        label: Text(label, maxLines: 1, softWrap: false),
                      ),
                  ],
                  selected: {_easyDays[i]},
                  onSelectionChanged: (v) =>
                      setState(() => _easyDays[i] = v.single),
                ),
              ),
            ],
          ),
        ),
    ];
  }

  /// A dropdown choosing one of [options] (value to label).
  Widget _orderDropdown({
    required String key,
    required String label,
    String? helper,
    required String value,
    required Map<String, String> options,
    required void Function(String) onChanged,
  }) {
    return DropdownButtonFormField<String>(
      key: Key(key),
      initialValue: value,
      isExpanded: true,
      decoration: InputDecoration(
        labelText: label,
        helperText: helper,
        helperMaxLines: 2,
      ),
      items: [
        for (final o in options.entries)
          DropdownMenuItem(value: o.key, child: Text(o.value)),
      ],
      onChanged: (v) => setState(() => onChanged(v!)),
    );
  }
}

/// Formats FSRS parameters for the parameters field.
String formatWeights(List<double> weights) =>
    weights.map((w) => w.toString()).join(', ');

/// Parses the parameters field: empty, or 19 numbers separated by commas
/// or spaces. Throws a [FormatException] explaining what is wrong.
List<double> parseWeights(String text) {
  final weights = <double>[];
  for (final part in text.trim().split(RegExp(r'[\s,]+'))) {
    if (part.isEmpty) continue;
    final w = double.tryParse(part);
    if (w == null || !w.isFinite) {
      throw FormatException(appL10n.notANumber(part));
    }
    weights.add(w);
  }
  if (weights.isNotEmpty && weights.length != fsrsWeightCount) {
    throw FormatException(appL10n.fsrsCountError(fsrsWeightCount));
  }
  return weights;
}

/// The number of FSRS parameters.
const fsrsWeightCount = 19;

/// Formats [date] as the options store it, e.g. 2026-01-31.
String formatDate(DateTime date) =>
    '${date.year.toString().padLeft(4, '0')}-'
    '${date.month.toString().padLeft(2, '0')}-'
    '${date.day.toString().padLeft(2, '0')}';

/// What an optimization is doing, from its latest progress [event].
String optimizeStage(golang.OptimizeEvent? event) {
  final l = appL10n;
  if (event == null) return l.starting;
  final done = percent(event.total == 0 ? 0 : event.done / event.total);
  return switch (event.stage) {
    'reading' => l.readingReviewHistory,
    'training' => l.optimizingPercent(done),
    'checking' => l.checkingParametersPercent(done),
    _ => l.optimizing,
  };
}

/// What Optimize found, for the screen.
String optimizeMessage(golang.OptimizeResult r) {
  final l = appL10n;
  if (r.notEnoughReviews) return l.notEnoughReviews(r.minReviews, r.reviews);
  if (r.use) return l.newParametersFitted(r.reviews);
  if (r.better && r.healthChecked && !r.healthPassed) {
    return l.healthCheckFailed;
  }
  return l.parametersAlreadyBest(r.reviews);
}

/// The longest allowed maximum and minimum interval, in days.
const maxIntervalDays = 36500;

/// The highest allowed maximum answer seconds.
const maxAnswerSeconds = 9999;

/// The longest auto advance time for a side, in seconds.
const maxAutoAdvanceSeconds = 9999;

/// The Easy Days levels, by their value in [golang.DeckOptions.easyDays].
List<String> easyDayLevels(AppLocalizations l) => [
  l.easyNormal,
  l.easyReduced,
  l.easyMinimum,
];

/// The weekdays, Monday first, as [golang.DeckOptions.easyDays] lists them.
List<String> weekdayAbbreviations(AppLocalizations l) {
  ensureDateFormats();
  final format = DateFormat.E(l.localeName);
  // 2024-01-01 was a Monday.
  return [for (var i = 1; i <= 7; i++) format.format(DateTime(2024, 1, i))];
}

const _stepUnits = {'d': 86400, 'h': 3600, 'm': 60, 's': 1};

/// Parses steps such as `1m 10m 1d` into seconds. Throws a
/// [FormatException] explaining what is wrong.
List<int> parseSteps(String text) {
  final steps = <int>[];
  for (final part in text.trim().split(RegExp(r'\s+'))) {
    if (part.isEmpty) continue;
    final m = RegExp(r'^(\d+(?:\.\d+)?)([smhd])$').firstMatch(part);
    if (m == null) {
      throw FormatException(appL10n.notADelay(part));
    }
    final seconds = (double.parse(m[1]!) * _stepUnits[m[2]]!).round();
    if (seconds <= 0) {
      throw FormatException(appL10n.stepTooShort);
    }
    if (seconds > 365 * 86400) {
      throw FormatException(appL10n.stepTooLong);
    }
    steps.add(seconds);
  }
  if (steps.length > 20) {
    throw FormatException(appL10n.tooManySteps);
  }
  return steps;
}

/// Formats steps in seconds as `1m 10m 1d`, each in its largest whole unit.
String formatSteps(List<int> steps) => steps.map(_formatStep).join(' ');

String _formatStep(int seconds) {
  final MapEntry(key: unit, value: size) = _stepUnits.entries.firstWhere(
    (u) => seconds % u.value == 0,
  );
  return '${seconds ~/ size}$unit';
}
