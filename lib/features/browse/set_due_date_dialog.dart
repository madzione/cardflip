import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Asks for the due date of [cards] cards and sets it with [apply], which
/// passes the input to Go (`golang.browse.setDueDate` or
/// `golang.study.setDueDate`). The input starts with the last one used. When
/// Go rejects it, its explanation shows under the field and the dialog stays
/// open. Returns what [apply] returned, or null when cancelled.
Future<R?> askSetDueDate<R extends Object>(
  BuildContext context, {
  required int cards,
  required Future<R> Function(String spec) apply,
}) => showDialog<R>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => _SetDueDateDialog<R>(
    cards: cards,
    initial: AppScope.read(context).settings.lastDueDate,
    apply: apply,
  ),
);

class _SetDueDateDialog<R extends Object> extends StatefulWidget {
  const _SetDueDateDialog({
    required this.cards,
    required this.initial,
    required this.apply,
  });

  final int cards;
  final String initial;
  final Future<R> Function(String spec) apply;

  @override
  State<_SetDueDateDialog<R>> createState() => _SetDueDateDialogState<R>();
}

class _SetDueDateDialogState<R extends Object>
    extends State<_SetDueDateDialog<R>> {
  late final _input = TextEditingController(text: widget.initial)
    ..selection = TextSelection(
      baseOffset: 0,
      extentOffset: widget.initial.length,
    );
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      final result = await widget.apply(_input.text);
      if (mounted) Navigator.pop<R>(context, result);
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = errorMessage(e);
          _busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final help = Theme.of(context).textTheme.bodySmall
        ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant);
    return AlertDialog(
      title: Text(
        widget.cards == 1
            ? context.l10n.setDueDate
            : context.l10n.setDueDateOf(widget.cards),
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          TextField(
            key: const Key('due-date-input'),
            controller: _input,
            autofocus: true,
            decoration: InputDecoration(
              labelText: context.l10n.daysFromToday,
              hintText: context.l10n.dueDateHint,
              errorText: _error,
              errorMaxLines: 3,
            ),
            onChanged: (_) {
              if (_error != null) setState(() => _error = null);
            },
            onSubmitted: (_) => _submit(),
          ),
          const SizedBox(height: 12),
          Text(context.l10n.dueDateHelp, style: help),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: _busy ? null : _submit,
          child: Text(context.l10n.setDueDate),
        ),
      ],
    );
  }
}
