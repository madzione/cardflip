import 'package:flutter/material.dart';
import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/l10n/l10n.dart';

/// What the reposition dialog asks for; see `golang.browse.reposition`.
typedef RepositionRequest = ({int start, int step, bool randomize, bool shift});

/// How a new card's due date reads: its position in the new card queue,
/// e.g. `New #42`, or [fallback] without one.
String newCardText(int? position, String fallback) =>
    position == null ? fallback : appL10n.newPosition(position);

/// The highest start position and step Go accepts.
const maxPosition = 1000000000;

/// Asks how to reposition the selected new cards. Returns null when
/// cancelled.
Future<RepositionRequest?> askReposition(BuildContext context) =>
    showDialog<RepositionRequest>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => const _RepositionDialog(),
    );

class _RepositionDialog extends StatefulWidget {
  const _RepositionDialog();

  @override
  State<_RepositionDialog> createState() => _RepositionDialogState();
}

class _RepositionDialogState extends State<_RepositionDialog> {
  final _form = GlobalKey<FormState>();
  final _start = TextEditingController(text: '0');
  final _step = TextEditingController(text: '1');
  bool _randomize = false;
  bool _shift = false;

  @override
  void dispose() {
    _start.dispose();
    _step.dispose();
    super.dispose();
  }

  /// A validator for a whole number from [min] to [maxPosition].
  FormFieldValidator<String> _number(int min) => (v) {
    final n = int.tryParse(v?.trim() ?? '');
    if (n == null || n < min || n > maxPosition) {
      return context.l10n.wholeNumberRange(min, maxPosition);
    }
    return null;
  };

  void _submit() {
    if (!_form.currentState!.validate()) return;
    Navigator.pop<RepositionRequest>(context, (
      start: int.parse(_start.text.trim()),
      step: int.parse(_step.text.trim()),
      randomize: _randomize,
      shift: _shift,
    ));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(context.l10n.repositionNewCards),
      content: Form(
        key: _form,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextFormField(
              key: const Key('reposition-start'),
              controller: _start,
              autofocus: true,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(
                labelText: context.l10n.startPosition,
              ),
              validator: _number(0),
              onFieldSubmitted: (_) => _submit(),
            ),
            TextFormField(
              key: const Key('reposition-step'),
              controller: _step,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(labelText: context.l10n.step),
              validator: _number(1),
              onFieldSubmitted: (_) => _submit(),
            ),
            const SizedBox(height: 8),
            CheckboxListTile(
              key: const Key('reposition-randomize'),
              contentPadding: EdgeInsets.zero,
              title: Text(context.l10n.randomizeOrder),
              value: _randomize,
              onChanged: (v) => setState(() => _randomize = v!),
            ),
            CheckboxListTile(
              key: const Key('reposition-shift'),
              contentPadding: EdgeInsets.zero,
              title: Text(context.l10n.shiftExisting),
              value: _shift,
              onChanged: (v) => setState(() => _shift = v!),
            ),
            Text(
              context.l10n.repositionHelp,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(onPressed: _submit, child: Text(context.l10n.reposition)),
      ],
    );
  }
}
