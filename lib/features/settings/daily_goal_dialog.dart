import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/l10n/l10n.dart';

/// The largest daily goal, in cards or minutes.
const maxDailyGoal = 9999;

/// The daily goal as settings show it: "30 cards a day", or "None".
String dailyGoalLabel(AppLocalizations l, String kind, int value) =>
    switch (kind) {
      'cards' => l.dailyGoalCardsPerDay(value),
      'minutes' => l.dailyGoalMinutesPerDay(value),
      _ => l.dailyGoalNone,
    };

/// Asks for the daily goal (none, a number of cards or of minutes) and
/// saves it.
Future<void> editDailyGoal(BuildContext context) => showDialog<void>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => const _DailyGoalDialog(),
);

class _DailyGoalDialog extends StatefulWidget {
  const _DailyGoalDialog();

  @override
  State<_DailyGoalDialog> createState() => _DailyGoalDialogState();
}

class _DailyGoalDialogState extends State<_DailyGoalDialog> {
  late final _settings = AppScope.read(context).settings;
  late String _kind = _settings.dailyGoalKind;
  late final _value = TextEditingController(
    text: _settings.dailyGoalKind.isEmpty
        ? '20'
        : _settings.dailyGoal.toString(),
  );
  bool _invalid = false;

  @override
  void dispose() {
    _value.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final value = int.tryParse(_value.text.trim()) ?? 0;
    if (_kind.isNotEmpty && (value < 1 || value > maxDailyGoal)) {
      setState(() => _invalid = true);
      return;
    }
    if (await _settings.setDailyGoal(_kind, _kind.isEmpty ? 0 : value) &&
        mounted) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    return AlertDialog(
      scrollable: true,
      title: Text(l.dailyGoal),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SegmentedButton<String>(
            key: const Key('daily-goal-kind'),
            showSelectedIcon: false,
            segments: [
              ButtonSegment(value: '', label: Text(l.dailyGoalNone)),
              ButtonSegment(value: 'cards', label: Text(l.dailyGoalCards)),
              ButtonSegment(value: 'minutes', label: Text(l.dailyGoalMinutes)),
            ],
            selected: {_kind},
            onSelectionChanged: (s) => setState(() {
              _kind = s.first;
              _invalid = false;
            }),
          ),
          if (_kind.isNotEmpty) ...[
            const SizedBox(height: 16),
            TextField(
              key: const Key('daily-goal-value'),
              controller: _value,
              autofocus: true,
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              decoration: InputDecoration(
                labelText: _kind == 'cards'
                    ? l.dailyGoalCardsLabel
                    : l.dailyGoalMinutesLabel,
                errorText: _invalid ? l.dailyGoalError : null,
              ),
              onSubmitted: (_) => _save(),
            ),
          ],
          const SizedBox(height: 12),
          Text(l.dailyGoalHelp, style: Theme.of(context).textTheme.bodySmall),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(l.cancel),
        ),
        FilledButton(onPressed: _save, child: Text(l.save)),
      ],
    );
  }
}
